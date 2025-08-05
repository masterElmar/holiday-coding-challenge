package models

import (
	"time"
)

// Hotel repräsentiert ein Hotel aus der hotels.csv
type Hotel struct {
	ID    int     `csv:"hotelid" json:"id"`
	Name  string  `csv:"hotelname" json:"name"`
	Stars float64 `csv:"hotelstars" json:"stars"`
}

// Offer repräsentiert ein Angebot für ein Hotel
type Offer struct {
	HotelID                   int       `csv:"hotelid" json:"hotelId"`
	OutboundDepartureDateTime time.Time `csv:"outbounddeparturedatetime" json:"outboundDepartureDateTime"`
	InboundDepartureDateTime  time.Time `csv:"inbounddeparturedatetime" json:"inboundDepartureDateTime"`
	CountAdults               int       `csv:"countadults" json:"countAdults"`
	CountChildren             int       `csv:"countchildren" json:"countChildren"`
	Price                     float64   `csv:"price" json:"price"`
	InboundDepartureAirport   string    `csv:"inbounddepartureairport" json:"inboundDepartureAirport"`
	InboundArrivalAirport     string    `csv:"inboundarrivalairport" json:"inboundArrivalAirport"`
	OutboundDepartureAirport  string    `csv:"outbounddepartureairport" json:"outboundDepartureAirport"`
	OutboundArrivalAirport    string    `csv:"outboundarrivalairport" json:"outboundArrivalAirport"`
}

// Duration berechnet die Dauer des Aufenthalts in Tagen
func (o *Offer) Duration() int {
	return int(o.InboundDepartureDateTime.Sub(o.OutboundDepartureDateTime).Hours() / 24)
}

// HotelWithBestOffer kombiniert Hotel-Informationen mit dem besten Angebot
type HotelWithBestOffer struct {
	Hotel     Hotel  `json:"hotel"`
	BestOffer *Offer `json:"bestOffer"`
}

// SearchParams repräsentiert die Such-Parameter
type SearchParams struct {
	DepartureAirports     []string  `query:"departureAirports"`
	EarliestDepartureDate time.Time `query:"earliestDepartureDate"`
	LatestReturnDate      time.Time `query:"latestReturnDate"`
	CountAdults           int       `query:"countAdults"`
	CountChildren         int       `query:"countChildren"`
	Duration              int       `query:"duration"`
}

// Matches prüft, ob ein Angebot den Such-Parametern entspricht
func (o *Offer) Matches(params SearchParams) bool {
	// Prüfe Departure Airports
	if len(params.DepartureAirports) > 0 {
		found := false
		for _, airport := range params.DepartureAirports {
			if o.OutboundDepartureAirport == airport {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Prüfe frühestes Abflugdatum
	if !params.EarliestDepartureDate.IsZero() && o.OutboundDepartureDateTime.Before(params.EarliestDepartureDate) {
		return false
	}

	// Prüfe spätestes Rückflugdatum
	if !params.LatestReturnDate.IsZero() && o.InboundDepartureDateTime.After(params.LatestReturnDate) {
		return false
	}

	// Prüfe Anzahl Erwachsene
	if params.CountAdults > 0 && o.CountAdults != params.CountAdults {
		return false
	}

	// Prüfe Anzahl Kinder
	if params.CountChildren > 0 && o.CountChildren != params.CountChildren {
		return false
	}

	// Prüfe Dauer
	if params.Duration > 0 && o.Duration() != params.Duration {
		return false
	}

	return true
}
