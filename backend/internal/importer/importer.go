package importer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"holiday-coding-challenge/backend/internal/models"

	"github.com/gocql/gocql"
)

// DataImporter lädt Daten aus CSV-Dateien
type DataImporter struct {
	hotelsPath string
	offersPath string
}

// NewDataImporter erstellt einen neuen DataImporter
func NewDataImporter(hotelsPath, offersPath string) *DataImporter {
	return &DataImporter{
		hotelsPath: hotelsPath,
		offersPath: offersPath,
	}
}

// LoadHotels lädt Hotel-Daten aus der CSV-Datei
func (d *DataImporter) LoadHotels() ([]models.Hotel, error) {
	file, err := os.Open(d.hotelsPath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Hotels-Datei: %w", err)
	}
	defer file.Close()

	// Manuell CSV mit Semikolon parsen
	reader := csv.NewReader(file)
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der CSV-Datei: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV-Datei ist leer")
	}

	// Header überspringen
	records = records[1:]

	var hotels []models.Hotel
	for i, record := range records {
		if len(record) < 3 {
			continue // Zeile überspringen, wenn nicht genug Spalten
		}

		hotel, err := parseHotelRecord(record)
		if err != nil {
			fmt.Printf("Warnung: Fehler beim Parsen von Zeile %d: %v\n", i+2, err)
			continue
		}

		hotels = append(hotels, hotel)
	}

	return hotels, nil
}

// LoadOffers lädt Angebots-Daten aus der CSV-Datei
func (d *DataImporter) LoadOffers() ([]models.Offer, error) {
	// Versuche erst die echte CSV-Datei zu laden
	offers, err := d.LoadOffersFromCSV()
	if err == nil {
		return offers, nil
	}
	// print error message
	fmt.Println("Warnung: Fehler beim Laden der Angebots-Datei:", err)

	// Falls die CSV-Datei nicht existiert oder fehlerhaft ist, verwende Beispiel-Daten
	fmt.Println("Warnung: Verwende Beispiel-Angebote, da CSV-Datei nicht geladen werden konnte")
	offers = d.generateSampleOffers()
	return offers, nil
}

// parseHotelRecord parst eine CSV-Zeile in ein Hotel-Objekt
func parseHotelRecord(record []string) (models.Hotel, error) {
	var hotel models.Hotel
	var err error

	// HotelID
	hotel.ID, err = strconv.Atoi(record[0])
	if err != nil {
		return hotel, fmt.Errorf("ungültige Hotel-ID: %s", record[0])
	}

	// Name
	hotel.Name = strings.TrimSpace(record[1])

	// Stars
	hotel.Stars, err = strconv.ParseFloat(record[2], 64)
	if err != nil {
		return hotel, fmt.Errorf("ungültige Sterne-Bewertung: %s", record[2])
	}

	return hotel, nil
}

// generateSampleOffers erstellt Beispiel-Angebote für Testing
func (d *DataImporter) generateSampleOffers() []models.Offer {
	now := time.Now()
	offers := []models.Offer{}

	// Erstelle Beispiel-Angebote für die ersten 10 Hotels
	for hotelID := 1; hotelID <= 10; hotelID++ {
		for i := 0; i < 5; i++ { // 5 Angebote pro Hotel
			departure := now.AddDate(0, 0, 7+i*7)    // Wöchentliche Abflüge
			returnDate := departure.AddDate(0, 0, 7) // 7 Tage Aufenthalt

			offer := models.Offer{
				HotelID:                  hotelID,
				DepartureDate:            departure,
				ReturnDate:               returnDate,
				CountAdults:              2,
				CountChildren:            0,
				Price:                    float64(800 + i*100 + hotelID*50), // Preisvariation
				InboundDepartureAirport:  "PMI",                             // Palma de Mallorca
				InboundArrivalAirport:    "FRA",                             // Frankfurt
				OutboundDepartureAirport: "FRA",                             // Frankfurt
				OutboundArrivalAirport:   "PMI",                             // Palma de Mallorca
			}
			offers = append(offers, offer)
		}
	}

	return offers
}

func (d *DataImporter) LoadOffersFromCSV() ([]models.Offer, error) {
	file, err := os.Open(d.offersPath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Angebots-Datei: %w", err)
	}
	defer file.Close()

	// 1MB Buffer für bessere I/O Performance
	bufferedReader := bufio.NewReaderSize(file, 1024*1024)
	reader := csv.NewReader(bufferedReader)
	reader.Comma = ','

	// Header überspringen
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen des Headers: %w", err)
	}

	// Kleinere Batches für große Dateien
	batchSize := 100
	numWorkers := runtime.NumCPU() * 2 // Mehr Worker

	batchChan := make(chan [][]string, 1) // Kleinerer Buffer
	resultChan := make(chan []models.Offer, numWorkers)
	errorChan := make(chan error, 10)
	progressChan := make(chan int, 100)

	var wg sync.WaitGroup

	// Progress Monitor hinzufügen
	go func() {
		processed := 0
		for count := range progressChan {
			processed += count
			if processed%10000 == 0 {
				fmt.Printf("Verarbeitet: %d Zeilen\n", processed)
			}
		}
	}()

	// Ihre bestehenden Worker
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				offers := d.processBatchParallel(batch, errorChan)
				if len(offers) > 0 {
					resultChan <- offers
				}
				progressChan <- len(batch) // Progress Update
			}
		}()
	}

	// Ihr bestehender Streaming-Code mit kleinen Änderungen
	totalRecords := 0
	go func() {
		defer close(batchChan)
		defer close(progressChan)
		batch := make([][]string, 0, batchSize)

		for {
			record, err := reader.Read()
			if err != nil {
				if len(batch) > 0 {
					batchChan <- batch
				}
				fmt.Printf("Gesamtanzahl Zeilen: %d\n", totalRecords)
				break
			}

			totalRecords++
			if len(record) >= 15 { // vollständiger Offers-Datensatz
				batch = append(batch, record)
			}

			if len(batch) >= batchSize {
				batchChan <- batch
				batch = make([][]string, 0, batchSize)
			}
		}
	}()

	// Rest bleibt gleich...
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	var allOffers []models.Offer
	for offerBatch := range resultChan {
		allOffers = append(allOffers, offerBatch...)
		// Regelmäßige Garbage Collection
		if len(allOffers)%50000 == 0 {
			runtime.GC()
		}
	}

	errorCount := 0
	for err := range errorChan {
		if errorCount < 10 {
			fmt.Printf("Warnung: Parsing-Fehler: %v\n", err)
		}
		errorCount++
	}

	return allOffers, nil
}

func (d *DataImporter) processBatchParallel(batch [][]string, errorChan chan<- error) []models.Offer {
	offers := make([]models.Offer, 0, len(batch))

	for _, record := range batch {
		offer, err := parseOfferRecord(record)
		if err != nil {
			select {
			case errorChan <- err:
			default:
			}
			continue
		}
		offers = append(offers, offer)
	}

	return offers
}

// parseOfferRecord parst eine CSV-Zeile in ein Offer-Objekt
func parseOfferRecord(record []string) (models.Offer, error) {
	var offer models.Offer
	var err error

	if len(record) < 15 {
		return offer, fmt.Errorf("zu wenige Spalten: %d", len(record))
	}

	// HotelID
	offer.HotelID, err = strconv.Atoi(strings.TrimSpace(record[0]))
	if err != nil {
		return offer, fmt.Errorf("ungültige Hotel-ID: %s", record[0])
	}

	// OutboundDepartureDateTime (trip start)
	offer.DepartureDate, err = parseDateTime(record[1])
	if err != nil {
		return offer, fmt.Errorf("ungültiges Abflugdatum: %s", record[1])
	}

	// InboundDepartureDateTime (trip end)
	offer.ReturnDate, err = parseDateTime(record[2])
	if err != nil {
		return offer, fmt.Errorf("ungültiges Rückflugdatum: %s", record[2])
	}

	// CountAdults
	offer.CountAdults, err = strconv.Atoi(strings.TrimSpace(record[3]))
	if err != nil {
		return offer, fmt.Errorf("ungültige Anzahl Erwachsene: %s", record[3])
	}

	// CountChildren
	offer.CountChildren, err = strconv.Atoi(strings.TrimSpace(record[4]))
	if err != nil {
		return offer, fmt.Errorf("ungültige Anzahl Kinder: %s", record[4])
	}

	// Price
	offer.Price, err = strconv.ParseFloat(strings.TrimSpace(record[5]), 64)
	if err != nil {
		return offer, fmt.Errorf("ungültiger Preis: %s", record[5])
	}

	// Airports & arrival datetimes
	offer.InboundDepartureAirport = strings.TrimSpace(record[6])
	offer.InboundArrivalAirport = strings.TrimSpace(record[7])
	if t, err := parseDateTime(record[8]); err == nil {
		offer.InboundArrivalDateTime = t
	} else {
		return offer, fmt.Errorf("ungültige inbound arrival datetime: %s", record[8])
	}
	offer.OutboundDepartureAirport = strings.TrimSpace(record[9])
	offer.OutboundArrivalAirport = strings.TrimSpace(record[10])
	if t, err := parseDateTime(record[11]); err == nil {
		offer.OutboundArrivalDateTime = t
	} else {
		return offer, fmt.Errorf("ungültige outbound arrival datetime: %s", record[11])
	}

	// Optional fields
	offer.MealType = strings.TrimSpace(record[12])
	offer.OceanView = parseBool(strings.TrimSpace(record[13]))
	offer.RoomType = strings.TrimSpace(record[14])

	return offer, nil
}

// parseDateTime parst einen DateTime-String in verschiedenen Formaten
func parseDateTime(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)

	// Verschiedene Datetime-Formate versuchen
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"02.01.2006 15:04:05",
		"02.01.2006",
		"2006-01-02",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unbekanntes Datetime-Format: %s", dateStr)
}

// parseBool parses common boolean string representations
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "y"
}

// ImportOffersToScylla streams the offers CSV and writes rows into Scylla using gocql
func (d *DataImporter) ImportOffersToScylla(session *gocql.Session) error {
	file, err := os.Open(d.offersPath)
	if err != nil {
		return fmt.Errorf("fehler beim Öffnen der Angebots-Datei: %w", err)
	}
	defer file.Close()

	// Größerer Buffer für schnelleres CSV-Streaming
	bufferedReader := bufio.NewReaderSize(file, 8*1024*1024)
	reader := csv.NewReader(bufferedReader)
	reader.Comma = ','

	// skip header
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("fehler beim Lesen des Headers: %w", err)
	}

	insertCQL := `INSERT INTO offers (
 		hotelid,
 		outbounddeparturedatetime,
 		inbounddeparturedatetime,
 		countadults,
 		countchildren,
 		price,
 		inbounddepartureairport,
 		inboundarrivalairport,
 		inboundarrivaldatetime,
 		outbounddepartureairport,
 		outboundarrivalairport,
 		outboundarrivaldatetime,
 		mealtype,
 		oceanview,
 		roomtype
 	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	// Hinweis: Die verwendete gocql-Version stellt hier kein Session.Prepare() bereit.
	// Wir erstellen Queries pro Write, markieren sie als idempotent und setzen Consistency auf ONE.

	jobs := make(chan []string, 2048)
	errs := make(chan error, 128)
	var wg sync.WaitGroup

	// Mehrere Worker für Parallelität; per Env IMPORT_WORKERS überschreibbar
	numWorkers := runtime.NumCPU() * 4
	if v := os.Getenv("IMPORT_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			numWorkers = n
		}
	}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rec := range jobs {
				o, err := parseOfferRecord(rec)
				if err != nil {
					select {
					case errs <- err:
					default:
					}
					continue
				}
				if err := session.Query(insertCQL,
					o.HotelID,
					o.DepartureDate,
					o.ReturnDate,
					o.CountAdults,
					o.CountChildren,
					o.Price,
					o.InboundDepartureAirport,
					o.InboundArrivalAirport,
					o.InboundArrivalDateTime,
					o.OutboundDepartureAirport,
					o.OutboundArrivalAirport,
					o.OutboundArrivalDateTime,
					o.MealType,
					o.OceanView,
					o.RoomType,
				).Consistency(gocql.One).Idempotent(true).Exec(); err != nil {
					select {
					case errs <- err:
					default:
					}
				}
			}
		}()
	}

	// producer
	go func() {
		count := 0
		for {
			rec, err := reader.Read()
			if err != nil {
				close(jobs)
				return
			}
			if len(rec) < 15 {
				continue
			}
			jobs <- rec
			count++
			if count%10000 == 0 {
				fmt.Printf("Import fortschritt: %d Zeilen geschrieben\n", count)
			}
		}
	}()

	// wait consumers
	wg.Wait()
	close(errs)

	// surface a few errors
	errCount := 0
	for e := range errs {
		if errCount < 10 {
			fmt.Printf("Warnung: Fehler beim Import: %v\n", e)
		}
		errCount++
	}
	if errCount > 0 {
		fmt.Printf("Gesamtzahl Importfehler: %d\n", errCount)
	}
	return nil
}
