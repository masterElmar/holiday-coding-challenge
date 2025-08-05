package config

import (
	"os"
	"strconv"
)

// Config enthält die Anwendungskonfiguration
type Config struct {
	Port           string
	HotelsDataPath string
	OffersDataPath string
}

// Load lädt die Konfiguration aus Umgebungsvariablen
func Load() *Config {
	config := &Config{
		Port:           getEnv("PORT", "8090"),
		HotelsDataPath: getEnv("HOTELS_DATA_PATH", "../data/hotels.csv"),
		OffersDataPath: getEnv("OFFERS_DATA_PATH", "../data/offers.csv"),
	}
	return config
}

// getEnv gibt den Wert einer Umgebungsvariablen zurück oder einen Standardwert
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gibt den Wert einer Umgebungsvariablen als Integer zurück oder einen Standardwert
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
