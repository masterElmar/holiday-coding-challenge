package storage

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"holiday-coding-challenge/backend/internal/models"

	"github.com/gocql/gocql"
)

// NewScyllaSession creates a gocql session using env or provided parameters.
func NewScyllaSession() (*gocql.Session, error) {
	hosts := getEnv("SCYLLA_HOSTS", "localhost")
	port := getEnvInt("SCYLLA_PORT", 9042)
	keyspace := getEnv("SCYLLA_KEYSPACE", "holidays")
	username := os.Getenv("SCYLLA_USERNAME")
	password := os.Getenv("SCYLLA_PASSWORD")
	consistencyEnv := strings.ToUpper(getEnv("SCYLLA_CONSISTENCY", "QUORUM"))
	dc := getEnv("SCYLLA_LOCAL_DC", "")
	numConns := getEnvInt("SCYLLA_NUM_CONNS", 4)

	cluster := gocql.NewCluster(strings.Split(hosts, ",")...)
	cluster.Port = port
	cluster.Keyspace = keyspace
	cluster.Consistency = parseConsistency(consistencyEnv)
	cluster.ProtoVersion = 4
	cluster.Timeout = 15 * time.Second
	cluster.ConnectTimeout = 15 * time.Second
	cluster.NumConns = numConns
	// Token-aware + optional DC-aware policy keeps requests close to data
	if dc != "" {
		cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.DCAwareRoundRobinPolicy(dc))
	} else {
		cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	}
	// Scylla-specific niceties helpful for containers/Cloud
	cluster.DisableInitialHostLookup = true
	cluster.IgnorePeerAddr = true
	// Hinweis: Shard-aware Port (19042) kann per SCYLLA_PORT konfiguriert werden
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{NumRetries: 5, Min: 200 * time.Millisecond, Max: 3 * time.Second}
	if username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{Username: username, Password: password}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("create scylla session: %w", err)
	}
	return session, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// Storage defines the methods our handlers need. Implemented by ScyllaStorage.
type Storage interface {
	GetHotelsWithBestOffers(params models.SearchParams) []models.HotelWithBestOffer
	GetOffersByHotel(hotelID int, params models.SearchParams) []models.Offer
	GetHotel(hotelID int) (*models.Hotel, bool)
	GetAllHotels() []models.Hotel
	GetStats() map[string]interface{}
	// GetAvailableDepartureAirports returns unique outbound departure airport codes across all offers
	GetAvailableDepartureAirports() []string
}

// ScyllaStorage implements Storage backed by ScyllaDB.
type ScyllaStorage struct {
	session *gocql.Session
}

func NewScyllaStorage(session *gocql.Session) *ScyllaStorage {
	return &ScyllaStorage{session: session}
}

// GetHotel returns a hotel by ID
func (s *ScyllaStorage) GetHotel(hotelID int) (*models.Hotel, bool) {
	var h models.Hotel
	var starsF32 float32
	q := `SELECT hotelid, hotelname, hotelstars FROM hotels WHERE hotelid = ?`
	if err := s.session.Query(q, hotelID).Consistency(gocql.One).Scan(&h.ID, &h.Name, &starsF32); err != nil {
		return nil, false
	}
	h.Stars = float64(starsF32)
	return &h, true
}

// GetAllHotels returns all hotels (small table)
func (s *ScyllaStorage) GetAllHotels() []models.Hotel {
	q := `SELECT hotelid, hotelname, hotelstars FROM hotels`
	iter := s.session.Query(q).Consistency(gocql.One).Iter()
	var (
		id       int
		name     string
		starsF32 float32
		res      []models.Hotel
	)
	for iter.Scan(&id, &name, &starsF32) {
		res = append(res, models.Hotel{ID: id, Name: name, Stars: float64(starsF32)})
	}
	_ = iter.Close()
	// keep deterministic order
	sort.Slice(res, func(i, j int) bool { return res[i].ID < res[j].ID })
	return res
}

// GetOffersByHotel fetches offers for a hotel and applies filters client-side for non-key attrs
func (s *ScyllaStorage) GetOffersByHotel(hotelID int, params models.SearchParams) []models.Offer {
	// Base: partition by hotel, rely on clustering by price ASC
	iter := s.offersIterByHotel(hotelID)
	var (
		o   models.Offer
		res []models.Offer
	)
	for {
		var ok bool
		o, ok = scanOffer(iter)
		if !ok {
			break
		}
		// Apply all filters using existing model logic
		if o.Matches(params) {
			res = append(res, o)
		}
	}
	_ = iter.Close()
	return res
}

// GetHotelsWithBestOffers returns hotels with their cheapest matching offer
func (s *ScyllaStorage) GetHotelsWithBestOffers(params models.SearchParams) []models.HotelWithBestOffer {
	hotels := s.GetAllHotels()
	println("Found", len(hotels), "hotels, searching for best offers...")
	results := make([]models.HotelWithBestOffer, 0, len(hotels))
	for _, h := range hotels {
		// Scan partition ordered by price (clustering), stop at first match
		iter := s.offersIterByHotel(h.ID)
		for {
			offer, ok := scanOffer(iter)
			if !ok {
				break
			}
			if offer.Matches(params) {
				results = append(results, models.HotelWithBestOffer{Hotel: h, BestOffer: &offer})
				break
			}
		}
		_ = iter.Close()
	}

	println("Found", len(results), "hotels with best offers")

	// sort by cheapest price
	sort.Slice(results, func(i, j int) bool {
		if results[i].BestOffer == nil {
			return false
		}
		if results[j].BestOffer == nil {
			return true
		}
		return results[i].BestOffer.Price < results[j].BestOffer.Price
	})
	return results
}

// GetStats returns simple stats. Note: COUNT(*) on large tables can be expensive.
func (s *ScyllaStorage) GetStats() map[string]interface{} {
	stats := map[string]interface{}{}
	var hotelsCount int64
	if err := s.session.Query(`SELECT COUNT(*) FROM hotels`).Scan(&hotelsCount); err == nil {
		stats["hotels"] = hotelsCount
	}
	var offersCount int64
	if err := s.session.Query(`SELECT COUNT(*) FROM offers`).Scan(&offersCount); err == nil {
		stats["offers"] = offersCount
	}
	// hotels_with_offers: cheap per-partition existence check
	hotels := s.GetAllHotels()
	withOffers := 0
	for _, h := range hotels {
		var price float64
		if err := s.session.Query(`SELECT price FROM offers WHERE hotelid = ? LIMIT 1`, h.ID).Consistency(gocql.One).Scan(&price); err == nil {
			withOffers++
		}
	}
	stats["hotels_with_offers"] = withOffers
	return stats
}

// --- Helpers & small utilities ---

// offersSelect encapsulates the commonly used offers projection
const offersSelect = `SELECT hotelid, outbounddeparturedatetime, inbounddeparturedatetime, countadults, countchildren, price, inbounddepartureairport, inboundarrivalairport, inboundarrivaldatetime, outbounddepartureairport, outboundarrivalairport, outboundarrivaldatetime, mealtype, oceanview, roomtype FROM offers WHERE hotelid = ?`

// offersIterByHotel creates an iterator over offers for a given hotel id with consistent settings
func (s *ScyllaStorage) offersIterByHotel(hotelID int) *gocql.Iter {
	return s.session.Query(offersSelect, hotelID).Consistency(gocql.One).Iter()
}

// GetAvailableDepartureAirports collects distinct outbound departure airport codes from all offers.
// Implementation scans offers partition per hotel; acceptable for demo/small datasets.
func (s *ScyllaStorage) GetAvailableDepartureAirports() []string {
	hotels := s.GetAllHotels()
	set := make(map[string]struct{})
	for _, h := range hotels {
		iter := s.offersIterByHotel(h.ID)
		for {
			offer, ok := scanOffer(iter)
			if !ok {
				break
			}
			if offer.OutboundDepartureAirport != "" {
				set[offer.OutboundDepartureAirport] = struct{}{}
			}
		}
		_ = iter.Close()
	}
	// to slice & sort
	res := make([]string, 0, len(set))
	for k := range set {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

// scanOffer reads the next row from iter and converts it to models.Offer
func scanOffer(iter *gocql.Iter) (models.Offer, bool) {
	var (
		hotelid                                                                      int
		outDep, inDep, inArr, outArr                                                 time.Time
		ca, cc                                                                       int
		price                                                                        float64
		inDepAirport, inArrAirport, outDepAirport, outArrAirport, mealType, roomType string
		oceanView                                                                    bool
	)
	if !iter.Scan(&hotelid, &outDep, &inDep, &ca, &cc, &price, &inDepAirport, &inArrAirport, &inArr, &outDepAirport, &outArrAirport, &outArr, &mealType, &oceanView, &roomType) {
		return models.Offer{}, false
	}
	return models.Offer{
		HotelID:                  hotelid,
		DepartureDate:            outDep,
		ReturnDate:               inDep,
		CountAdults:              ca,
		CountChildren:            cc,
		Price:                    price,
		InboundDepartureAirport:  inDepAirport,
		InboundArrivalAirport:    inArrAirport,
		InboundArrivalDateTime:   inArr,
		OutboundDepartureAirport: outDepAirport,
		OutboundArrivalAirport:   outArrAirport,
		OutboundArrivalDateTime:  outArr,
		MealType:                 mealType,
		OceanView:                oceanView,
		RoomType:                 roomType,
	}, true
}

// parseConsistency maps string to gocql.Consistency with sane default
func parseConsistency(v string) gocql.Consistency {
	switch strings.ToUpper(v) {
	case "ANY":
		return gocql.Any
	case "ONE":
		return gocql.One
	case "TWO":
		return gocql.Two
	case "THREE":
		return gocql.Three
	case "QUORUM":
		return gocql.Quorum
	case "ALL":
		return gocql.All
	case "LOCAL_QUORUM":
		return gocql.LocalQuorum
	case "EACH_QUORUM":
		return gocql.EachQuorum
	case "LOCAL_ONE":
		return gocql.LocalOne
	default:
		return gocql.Quorum
	}
}
