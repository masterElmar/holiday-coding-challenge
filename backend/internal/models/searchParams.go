package models

import "time"

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
	return o.matchesDepartureAirports(params.DepartureAirports) &&
		o.matchesEarliestDepartureDate(params.EarliestDepartureDate) &&
		o.matchesLatestReturnDate(params.LatestReturnDate) &&
		o.matchesCountAdults(params.CountAdults) &&
		o.matchesCountChildren(params.CountChildren) &&
		o.matchesDuration(params.Duration)
}

// matchesDepartureAirports prüft, ob das Angebot einen der gewünschten Abflughäfen hat
func (o *Offer) matchesDepartureAirports(airports []string) bool {
	if len(airports) == 0 {
		return true
	}
	for _, airport := range airports {
		if o.OutboundDepartureAirport == airport {
			return true
		}
	}
	return false
}

// matchesEarliestDepartureDate prüft das früheste Abflugdatum
func (o *Offer) matchesEarliestDepartureDate(earliestDate time.Time) bool {
	return earliestDate.IsZero() || !o.DepartureDate.Before(earliestDate)
}

// matchesLatestReturnDate prüft das späteste Rückflugdatum
func (o *Offer) matchesLatestReturnDate(latestDate time.Time) bool {
	return latestDate.IsZero() || !o.ReturnDate.After(latestDate)
}

// matchesCountAdults prüft die Anzahl Erwachsene
func (o *Offer) matchesCountAdults(countAdults int) bool {
	return countAdults == 0 || o.CountAdults == countAdults
}

// matchesCountChildren prüft die Anzahl Kinder
func (o *Offer) matchesCountChildren(countChildren int) bool {
	return countChildren == 0 || o.CountChildren == countChildren
}

// matchesDuration prüft die Dauer
func (o *Offer) matchesDuration(duration int) bool {
	return duration == 0 || o.Duration() == duration
}
