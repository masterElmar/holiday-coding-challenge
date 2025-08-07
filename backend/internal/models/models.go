package models

import (
	"time"
)

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