package main

import (
	"fmt"
	"log"
	"time"

	"holiday-coding-challenge/backend/internal/config"
	"holiday-coding-challenge/backend/internal/handlers"
	"holiday-coding-challenge/backend/internal/importer"
	"holiday-coding-challenge/backend/internal/storage"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Konfiguration laden
	cfg := config.Load()

	// Fiber App erstellen
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// Middleware hinzufügen
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Storage initialisieren
	memStorage := storage.NewMemoryStorage()

	// Daten laden
	fmt.Println("Lade Hotel-Daten...")
	dataImporter := importer.NewDataImporter(cfg.HotelsDataPath, cfg.OffersDataPath)

	hotels, err := dataImporter.LoadHotels()
	if err != nil {
		log.Fatalf("Fehler beim Laden der Hotel-Daten: %v", err)
	}
	memStorage.LoadHotels(hotels)
	fmt.Printf("✓ %d Hotels geladen\n", len(hotels))

	fmt.Println("Lade Angebots-Daten...")
	offers, err := dataImporter.LoadOffers()
	if err != nil {
		log.Fatalf("Fehler beim Laden der Angebots-Daten: %v", err)
	}
	memStorage.LoadOffers(offers)
	fmt.Printf("✓ %d Angebote geladen\n", len(offers))

	// Handler initialisieren
	hotelHandler := handlers.NewHotelHandler(memStorage)

	// Huma API konfigurieren
	config := huma.DefaultConfig("Holiday Coding Challenge API", "1.0.0")
	config.OpenAPI.Info.Description = "API für Hotel-Suche und Angebote"
	config.OpenAPI.Servers = []*huma.Server{
		{URL: "http://localhost:8090", Description: "Development server"},
	}
	api := humafiber.New(app, config)

	// Huma API-Routen definieren
	huma.Register(api, huma.Operation{
		OperationID: "getBestOffersByHotel",
		Method:      "GET",
		Path:        "/bestOffersByHotel",
		Summary:     "Get best offers by hotel",
		Description: "Get the best (i.e. cheapest) offer for every hotel that has at least one available offer for a given search",
		Tags:        []string{"hotels"},
	}, hotelHandler.HumaGetHotelsWithBestOffers)

	huma.Register(api, huma.Operation{
		OperationID: "GetHotelOffers",
		Method:      "GET",
		Path:        "/hotels/{hotelId}/offers",
		Summary:     "Get hotel offers",
		Description: "Get available offers for a given hotel",
		Tags:        []string{"hotels", "offers"},
	}, hotelHandler.HumaGetOffersByHotel)

	huma.Register(api, huma.Operation{
		OperationID: "getStats",
		Method:      "GET",
		Path:        "/api/stats",
		Summary:     "Get statistics",
		Description: "Retrieve data statistics",
		Tags:        []string{"stats"},
	}, hotelHandler.HumaGetStats)

	// Health Check
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// Root-Route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Holiday Coding Challenge Backend API",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /api/health - Health Check",
				"GET /api/stats - Datenstatistiken",
				"GET /api/hotels - Hotels mit besten Angeboten",
				"GET /api/hotels/{id}/offers - Alle Angebote für ein Hotel",
				"GET /docs - OpenAPI Documentation",
			},
			"example_queries": []string{
				"/api/hotels?departureAirports=FRA,MUC&earliestDepartureDate=2025-08-10&latestReturnDate=2025-08-31&countAdults=2&countChildren=0&duration=7",
				"/api/hotels/1/offers?departureAirports=FRA&countAdults=2&countChildren=0",
			},
		})
	})

	// Server starten
	fmt.Printf("\n🚀 Server startet auf Port %s\n", cfg.Port)
	fmt.Printf("📍 API-Dokumentation: http://localhost:%s/docs\n", cfg.Port)
	fmt.Printf("📊 Statistiken: http://localhost:%s/api/stats\n", cfg.Port)
	fmt.Printf("🏨 Hotels: http://localhost:%s/api/hotels\n", cfg.Port)

	log.Fatal(app.Listen(":" + cfg.Port))
}
