package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"holiday-coding-challenge/backend/internal/models"
	"holiday-coding-challenge/backend/internal/storage"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
)

// HotelHandler behandelt Hotel-bezogene API-Anfragen
type HotelHandler struct {
	storage storage.Storage
}

// NewHotelHandler erstellt einen neuen HotelHandler
func NewHotelHandler(storage storage.Storage) *HotelHandler {
	return &HotelHandler{
		storage: storage,
	}
}

// HumaGetHotelsWithBestOffers - Huma-kompatible Version
func (h *HotelHandler) HumaGetHotelsWithBestOffers(ctx context.Context, input *struct {
	models.ApiSearchParams
}) (*models.BestOffersByHotelResponse, error) {
	params, err := h.convertSearchParams(input.ApiSearchParams)
	if err != nil {
		return nil, huma.Error400BadRequest("Ungültige Such-Parameter: " + err.Error())
	}

	hotels := h.storage.GetHotelsWithBestOffers(params)

	// Konvertiere zu Frontend-kompatiblem Format
	bestOffers := make([]models.BestHotelOffer, len(hotels))
	for i, hotel := range hotels {
		bestOffers[i] = models.BestHotelOffer{
			Hotel:                hotel.Hotel,
			MinPrice:             hotel.BestOffer.Price,
			DepartureDate:        hotel.BestOffer.OutboundArrivalDateTime.Format("2006-01-02"),
			ReturnDate:           hotel.BestOffer.InboundArrivalDateTime.Format("2006-01-02"),
			RoomType:             "", // TODO: Falls verfügbar in Offer-Struktur
			MealType:             "", // TODO: Falls verfügbar in Offer-Struktur
			CountAdults:          hotel.BestOffer.CountAdults,
			CountChildren:        hotel.BestOffer.CountChildren,
			Duration:             hotel.BestOffer.Duration(),
			CountAvailableOffers: 1, // TODO: Tatsächliche Anzahl berechnen
		}
	}

	resp := &models.BestOffersByHotelResponse{}
	resp.Body = bestOffers

	return resp, nil
}

// HumaGetOffersByHotel - Huma-kompatible Version
func (h *HotelHandler) HumaGetOffersByHotel(ctx context.Context, input *struct {
	ID int `path:"hotelId" doc:"Hotel ID"`
	models.ApiSearchParams
}) (*models.HotelOffersResponse, error) {
	// Prüfen, ob das Hotel existiert
	hotel, exists := h.storage.GetHotel(input.ID)
	if !exists {
		return nil, huma.Error404NotFound("Hotel nicht gefunden")
	}

	// Such-Parameter konvertieren
	params, err := h.convertSearchParams(input.ApiSearchParams)
	if err != nil {
		return nil, huma.Error400BadRequest("Ungültige Such-Parameter: " + err.Error())
	}

	// Angebote für das Hotel abrufen
	offers := h.storage.GetOffersByHotel(input.ID, params)

	resp := &models.HotelOffersResponse{}
	resp.Body.Hotel = *hotel // Dereferenziere den Pointer
	resp.Body.Items = offers

	return resp, nil
}

// HumaGetStats - Huma-kompatible Version
func (h *HotelHandler) HumaGetStats(ctx context.Context, input *struct{}) (*models.StatsResponse, error) {
	stats := h.storage.GetStats()

	resp := &models.StatsResponse{}
	resp.Body = stats

	return resp, nil
}

// convertSearchParams konvertiert Huma SearchParams zu models.SearchParams
func (h *HotelHandler) convertSearchParams(params models.ApiSearchParams) (models.SearchParams, error) {
	var result models.SearchParams

	// Departure Airports
	result.DepartureAirports = params.DepartureAirports

	// Earliest Departure Date - unterstützt sowohl ISO-8601 DateTime als auch einfache Datums-Formate
	if params.EarliestDepartureDate != "" {
		var date time.Time
		var err error

		// Versuche zuerst ISO-8601 DateTime Format
		if date, err = time.Parse(time.RFC3339, params.EarliestDepartureDate); err != nil {
			// Fallback zu einfachem Datum
			if date, err = time.Parse("2006-01-02", params.EarliestDepartureDate); err != nil {
				return result, err
			}
		}
		result.EarliestDepartureDate = date
	}

	// Latest Return Date - unterstützt sowohl ISO-8601 DateTime als auch einfache Datums-Formate
	if params.LatestReturnDate != "" {
		var date time.Time
		var err error

		// Versuche zuerst ISO-8601 DateTime Format
		if date, err = time.Parse(time.RFC3339, params.LatestReturnDate); err != nil {
			// Fallback zu einfachem Datum
			if date, err = time.Parse("2006-01-02", params.LatestReturnDate); err != nil {
				return result, err
			}
		}
		result.LatestReturnDate = date
	}

	// Counts and Duration
	result.CountAdults = params.CountAdults
	result.CountChildren = params.CountChildren
	result.Duration = params.Duration

	return result, nil
}

// GetHotelsWithBestOffers gibt Hotels mit ihren besten Angeboten zurück
// GET /api/hotels?departureAirports=FRA,MUC&earliestDepartureDate=2025-08-10&latestReturnDate=2025-08-31&countAdults=2&countChildren=0&duration=7
func (h *HotelHandler) GetHotelsWithBestOffers(c *fiber.Ctx) error {
	params, err := h.parseSearchParams(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ungültige Such-Parameter: " + err.Error(),
		})
	}

	hotels := h.storage.GetHotelsWithBestOffers(params)

	return c.JSON(fiber.Map{
		"hotels": hotels,
		"total":  len(hotels),
		"params": params,
	})
}

// GetOffersByHotel gibt alle Angebote für ein bestimmtes Hotel zurück
// GET /api/hotels/:id/offers?departureAirports=FRA&earliestDepartureDate=2025-08-10&latestReturnDate=2025-08-31&countAdults=2&countChildren=0&duration=7
func (h *HotelHandler) GetOffersByHotel(c *fiber.Ctx) error {
	// Hotel-ID aus der URL extrahieren
	hotelIDStr := c.Params("id")
	hotelID, err := strconv.Atoi(hotelIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ungültige Hotel-ID",
		})
	}

	// Prüfen, ob das Hotel existiert
	hotel, exists := h.storage.GetHotel(hotelID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Hotel nicht gefunden",
		})
	}

	// Such-Parameter parsen
	params, err := h.parseSearchParams(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ungültige Such-Parameter: " + err.Error(),
		})
	}

	// Angebote für das Hotel abrufen
	offers := h.storage.GetOffersByHotel(hotelID, params)

	return c.JSON(fiber.Map{
		"hotel":  hotel,
		"offers": offers,
		"total":  len(offers),
		"params": params,
	})
}

// GetStats gibt Statistiken über die geladenen Daten zurück
// GET /api/stats
func (h *HotelHandler) GetStats(c *fiber.Ctx) error {
	stats := h.storage.GetStats()
	return c.JSON(stats)
}

// parseSearchParams parst die Such-Parameter aus der Query
func (h *HotelHandler) parseSearchParams(c *fiber.Ctx) (models.SearchParams, error) {
	var params models.SearchParams

	// Departure Airports (komma-getrennt)
	if departureAirports := c.Query("departureAirports"); departureAirports != "" {
		airports := strings.Split(departureAirports, ",")
		for i, airport := range airports {
			airports[i] = strings.TrimSpace(airport)
		}
		params.DepartureAirports = airports
	}

	// Earliest Departure Date
	if earliestDate := c.Query("earliestDepartureDate"); earliestDate != "" {
		date, err := time.Parse("2006-01-02", earliestDate)
		if err != nil {
			return params, err
		}
		params.EarliestDepartureDate = date
	}

	// Latest Return Date
	if latestDate := c.Query("latestReturnDate"); latestDate != "" {
		date, err := time.Parse("2006-01-02", latestDate)
		if err != nil {
			return params, err
		}
		params.LatestReturnDate = date
	}

	// Count Adults
	if countAdults := c.Query("countAdults"); countAdults != "" {
		count, err := strconv.Atoi(countAdults)
		if err != nil {
			return params, err
		}
		params.CountAdults = count
	}

	// Count Children
	if countChildren := c.Query("countChildren"); countChildren != "" {
		count, err := strconv.Atoi(countChildren)
		if err != nil {
			return params, err
		}
		params.CountChildren = count
	}

	// Duration
	if duration := c.Query("duration"); duration != "" {
		dur, err := strconv.Atoi(duration)
		if err != nil {
			return params, err
		}
		params.Duration = dur
	}

	return params, nil
}
