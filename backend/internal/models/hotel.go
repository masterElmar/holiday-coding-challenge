package models

// Hotel repräsentiert ein Hotel aus der hotels.csv
type Hotel struct {
	ID    int     `csv:"hotelid" json:"id"`
	Name  string  `csv:"hotelname" json:"name"`
	Stars float64 `csv:"hotelstars" json:"stars"`
}

// HotelWithBestOffer kombiniert Hotel-Informationen mit dem besten Angebot
type HotelWithBestOffer struct {
	Hotel     Hotel  `json:"hotel"`
	BestOffer *Offer `json:"bestOffer"`
}

// SearchParams für Huma API
type ApiSearchParams struct {
	DepartureAirports     []string `query:"departureAirports" doc:"Comma-separated list of departure airports (e.g., FRA,MUC)"`
	EarliestDepartureDate string   `query:"earliestDepartureDate" doc:"Earliest departure date (YYYY-MM-DD)"`
	LatestReturnDate      string   `query:"latestReturnDate" doc:"Latest return date (YYYY-MM-DD)"`
	CountAdults           int      `query:"countAdults" doc:"Number of adults"`
	CountChildren         int      `query:"countChildren" doc:"Number of children"`
	Duration              int      `query:"duration" doc:"Trip duration in days"`
}

// BestHotelOffer entspricht der Frontend-Erwartung
type BestHotelOffer struct {
	Hotel                Hotel   `json:"hotel"`
	MinPrice             float64 `json:"minPrice"`
	DepartureDate        string  `json:"departureDate"`
	ReturnDate           string  `json:"returnDate"`
	RoomType             string  `json:"roomType,omitempty"`
	MealType             string  `json:"mealType,omitempty"`
	CountAdults          int     `json:"countAdults"`
	CountChildren        int     `json:"countChildren"`
	Duration             int     `json:"duration"`
	CountAvailableOffers int     `json:"countAvailableOffers"`
}

// BestOffersByHotelResponse für Huma API - kompatibel mit Frontend
type BestOffersByHotelResponse struct {
	Body []BestHotelOffer `json:"body"`
}

// HotelOffersResponse für Huma API - kompatibel mit Frontend
type HotelOffersResponse struct {
	Body struct {
		Hotel Hotel   `json:"hotel"`
		Items []Offer `json:"items"`
	} `json:"body"`
}

// StatsResponse für Huma API
type StatsResponse struct {
	Body map[string]interface{} `json:"stats"`
}

// AirportsResponse for Huma API
type AirportsResponse struct {
	Body []string `json:"airports"`
}
