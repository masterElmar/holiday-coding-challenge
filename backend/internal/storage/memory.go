package storage

import (
	"sort"
	"sync"

	"holiday-coding-challenge/backend/internal/models"
)

// MemoryStorage speichert Hotels und Angebote im Arbeitsspeicher für schnelle Zugriffe
type MemoryStorage struct {
	hotels        map[int]*models.Hotel
	offers        []models.Offer
	offersByHotel map[int][]models.Offer
	mutex         sync.RWMutex
}

// NewMemoryStorage erstellt einen neuen MemoryStorage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		hotels:        make(map[int]*models.Hotel),
		offers:        make([]models.Offer, 0),
		offersByHotel: make(map[int][]models.Offer),
	}
}

// LoadHotels lädt Hotels in den Speicher
func (ms *MemoryStorage) LoadHotels(hotels []models.Hotel) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.hotels = make(map[int]*models.Hotel)
	for i := range hotels {
		ms.hotels[hotels[i].ID] = &hotels[i]
	}
}

// LoadOffers lädt Angebote in den Speicher und indexiert sie nach Hotel-ID
func (ms *MemoryStorage) LoadOffers(offers []models.Offer) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.offers = offers
	ms.offersByHotel = make(map[int][]models.Offer)

	for _, offer := range offers {
		ms.offersByHotel[offer.HotelID] = append(ms.offersByHotel[offer.HotelID], offer)
	}

	// Sortiere Angebote pro Hotel nach Preis (günstigstes zuerst)
	for hotelID := range ms.offersByHotel {
		sort.Slice(ms.offersByHotel[hotelID], func(i, j int) bool {
			return ms.offersByHotel[hotelID][i].Price < ms.offersByHotel[hotelID][j].Price
		})
	}
}

// GetHotelsWithBestOffers gibt Hotels mit ihren besten Angeboten zurück, die den Suchparametern entsprechen
func (ms *MemoryStorage) GetHotelsWithBestOffers(params models.SearchParams) []models.HotelWithBestOffer {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	var results []models.HotelWithBestOffer

	for hotelID, offers := range ms.offersByHotel {
		hotel, exists := ms.hotels[hotelID]
		if !exists {
			continue
		}

		// Finde das beste Angebot, das den Parametern entspricht
		var bestOffer *models.Offer
		for i := range offers {
			if offers[i].Matches(params) {
				bestOffer = &offers[i]
				break // Das erste Angebot ist das günstigste (bereits sortiert)
			}
		}

		// Nur Hotels mit passenden Angeboten zurückgeben
		if bestOffer != nil {
			results = append(results, models.HotelWithBestOffer{
				Hotel:     *hotel,
				BestOffer: bestOffer,
			})
		}
	}

	// Sortiere Ergebnisse nach Preis des besten Angebots
	sort.Slice(results, func(i, j int) bool {
		return results[i].BestOffer.Price < results[j].BestOffer.Price
	})

	return results
}

// GetOffersByHotel gibt alle Angebote für ein bestimmtes Hotel zurück, die den Suchparametern entsprechen
func (ms *MemoryStorage) GetOffersByHotel(hotelID int, params models.SearchParams) []models.Offer {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	offers, exists := ms.offersByHotel[hotelID]
	if !exists {
		return []models.Offer{}
	}

	var filteredOffers []models.Offer
	for _, offer := range offers {
		if offer.Matches(params) {
			filteredOffers = append(filteredOffers, offer)
		}
	}

	return filteredOffers
}

// GetHotel gibt ein Hotel nach ID zurück
func (ms *MemoryStorage) GetHotel(hotelID int) (*models.Hotel, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	hotel, exists := ms.hotels[hotelID]
	return hotel, exists
}

// GetAllHotels gibt alle Hotels zurück
func (ms *MemoryStorage) GetAllHotels() []models.Hotel {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	hotels := make([]models.Hotel, 0, len(ms.hotels))
	for _, hotel := range ms.hotels {
		hotels = append(hotels, *hotel)
	}

	// Sortiere nach Hotel-ID
	sort.Slice(hotels, func(i, j int) bool {
		return hotels[i].ID < hotels[j].ID
	})

	return hotels
}

// GetStats gibt Statistiken über die geladenen Daten zurück
func (ms *MemoryStorage) GetStats() map[string]interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	return map[string]interface{}{
		"hotels":             len(ms.hotels),
		"offers":             len(ms.offers),
		"hotels_with_offers": len(ms.offersByHotel),
	}
}
