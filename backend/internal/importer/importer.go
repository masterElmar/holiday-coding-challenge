package importer

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"holiday-coding-challenge/backend/internal/models"
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
	if offers, err := d.LoadOffersFromCSV(); err == nil {
		return offers, nil
	}

	// Falls die CSV-Datei nicht existiert oder fehlerhaft ist, verwende Beispiel-Daten
	fmt.Println("Warnung: Verwende Beispiel-Angebote, da CSV-Datei nicht geladen werden konnte")
	offers := d.generateSampleOffers()
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
				HotelID:                   hotelID,
				OutboundDepartureDateTime: departure,
				InboundDepartureDateTime:  returnDate,
				CountAdults:               2,
				CountChildren:             0,
				Price:                     float64(800 + i*100 + hotelID*50), // Preisvariation
				InboundDepartureAirport:   "PMI",                             // Palma de Mallorca
				InboundArrivalAirport:     "FRA",                             // Frankfurt
				OutboundDepartureAirport:  "FRA",                             // Frankfurt
				OutboundArrivalAirport:    "PMI",                             // Palma de Mallorca
			}
			offers = append(offers, offer)
		}
	}

	return offers
}

// LoadOffersFromCSV lädt Angebote aus einer echten CSV-Datei
func (d *DataImporter) LoadOffersFromCSV() ([]models.Offer, error) {
	file, err := os.Open(d.offersPath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Angebots-Datei: %w", err)
	}
	defer file.Close()

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

	var offers []models.Offer
	for i, record := range records {
		if len(record) < 10 {
			continue // Zeile überspringen, wenn nicht genug Spalten
		}

		offer, err := parseOfferRecord(record)
		if err != nil {
			fmt.Printf("Warnung: Fehler beim Parsen von Zeile %d: %v\n", i+2, err)
			continue
		}

		offers = append(offers, offer)
	}

	return offers, nil
}

// parseOfferRecord parst eine CSV-Zeile in ein Offer-Objekt
func parseOfferRecord(record []string) (models.Offer, error) {
	var offer models.Offer
	var err error

	// HotelID
	offer.HotelID, err = strconv.Atoi(record[0])
	if err != nil {
		return offer, fmt.Errorf("ungültige Hotel-ID: %s", record[0])
	}

	// OutboundDepartureDateTime
	offer.OutboundDepartureDateTime, err = parseDateTime(record[1])
	if err != nil {
		return offer, fmt.Errorf("ungültiges Abflugdatum: %s", record[1])
	}

	// InboundDepartureDateTime
	offer.InboundDepartureDateTime, err = parseDateTime(record[2])
	if err != nil {
		return offer, fmt.Errorf("ungültiges Rückflugdatum: %s", record[2])
	}

	// CountAdults
	offer.CountAdults, err = strconv.Atoi(record[3])
	if err != nil {
		return offer, fmt.Errorf("ungültige Anzahl Erwachsene: %s", record[3])
	}

	// CountChildren
	offer.CountChildren, err = strconv.Atoi(record[4])
	if err != nil {
		return offer, fmt.Errorf("ungültige Anzahl Kinder: %s", record[4])
	}

	// Price
	offer.Price, err = strconv.ParseFloat(record[5], 64)
	if err != nil {
		return offer, fmt.Errorf("ungültiger Preis: %s", record[5])
	}

	// Airports
	offer.InboundDepartureAirport = strings.TrimSpace(record[6])
	offer.InboundArrivalAirport = strings.TrimSpace(record[7])
	offer.OutboundDepartureAirport = strings.TrimSpace(record[8])
	offer.OutboundArrivalAirport = strings.TrimSpace(record[9])

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
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unbekanntes Datetime-Format: %s", dateStr)
}
