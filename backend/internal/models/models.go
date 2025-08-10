package models

import (
	"time"
)

// Offer repräsentiert ein Angebot für ein Hotel
type Offer struct {
	HotelID                  int       `csv:"hotelid" json:"hotelId"`
	DepartureDate            time.Time `csv:"departuredate" json:"departureDate"`
	ReturnDate               time.Time `csv:"returndate" json:"returnDate"`
	CountAdults              int       `csv:"countadults" json:"countAdults"`
	CountChildren            int       `csv:"countchildren" json:"countChildren"`
	Price                    float64   `csv:"price" json:"price"`
	InboundDepartureAirport  string    `csv:"inbounddepartureairport" json:"inboundDepartureAirport"`
	InboundArrivalAirport    string    `csv:"inboundarrivalairport" json:"inboundArrivalAirport"`
	InboundArrivalDateTime   time.Time `csv:"inboundarrivaldatetime" json:"inboundArrivalDateTime"`
	OutboundDepartureAirport string    `csv:"outbounddepartureairport" json:"outboundDepartureAirport"`
	OutboundArrivalAirport   string    `csv:"outboundarrivalairport" json:"outboundArrivalAirport"`
	OutboundArrivalDateTime  time.Time `csv:"outboundarrivaldatetime" json:"outboundArrivalDateTime"`
	MealType                 string    `csv:"mealtype,omitempty" json:"mealType,omitempty"`
	OceanView                bool      `csv:"oceanview,omitempty" json:"oceanView,omitempty"`
	RoomType                 string    `csv:"roomtype,omitempty" json:"roomType,omitempty"`
}

// Duration berechnet die Dauer des Aufenthalts in Tagen
func (o *Offer) Duration() int {
	return int(o.InboundArrivalDateTime.Sub(o.OutboundArrivalDateTime).Hours() / 24)
}
