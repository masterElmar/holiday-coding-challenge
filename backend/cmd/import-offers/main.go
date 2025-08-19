package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"holiday-coding-challenge/backend/internal/config"
	"holiday-coding-challenge/backend/internal/importer"
	"holiday-coding-challenge/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	// allow overriding offers path via flag
	offersPath := flag.String("offers", cfg.OffersDataPath, "Path to offers CSV")
	flag.Parse()

	// connect to scylla
	session, err := storage.NewScyllaSession()
	if err != nil {
		log.Fatalf("Scylla session error: %v", err)
	}
	defer session.Close()

	// ensure schema keyspace is active (handled by session setup keyspace)
	imp := importer.NewDataImporter("", *offersPath)
	start := time.Now()
	fmt.Printf("Starting offers import to Scylla from %s...\n", *offersPath)
	if err := imp.ImportOffersToScylla(session); err != nil {
		log.Fatalf("Import failed: %v", err)
	}
	fmt.Printf("Done in %s.\n", time.Since(start))
}
