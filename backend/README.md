# Holiday Challenge Backend

## Umgebungsvariablen

Erstelle eine `.env` Datei im Backend-Verzeichnis:

```env
# Server Configuration
PORT=8080
LOG_LEVEL=info
ENABLE_CORS=true

# Data Sources
HOTELS_FILE_PATH=../data/hotels.csv
OFFERS_FILE_PATH=../data/offers.csv

# Performance Settings
MAX_OFFERS_CACHED=1000000
```

## Verfügbare Umgebungsvariablen

| Variable | Beschreibung | Standard |
|----------|--------------|----------|
| `PORT` | Server Port | `8080` |
| `HOTELS_FILE_PATH` | Pfad zur Hotels CSV Datei | `../data/hotels.csv` |
| `OFFERS_FILE_PATH` | Pfad zur Angebote CSV Datei | `../data/offers.csv` |
| `LOG_LEVEL` | Log Level (debug, info, warn, error) | `info` |
| `ENABLE_CORS` | CORS aktivieren | `true` |
| `MAX_OFFERS_CACHED` | Maximale Anzahl gecachter Angebote | `1000000` |

## Installation

```bash
cd backend
go mod download
```

## Ausführen

```bash
cd backend
go run cmd/server/main.go
```

## Verfügbare Endpunkte

- `GET /health` - Gesundheitsstatus und Statistiken
- `GET /hotels` - Alle Hotels auflisten

## Datenstrukturen

### Hotel
```json
{
  "hotelid": 1,
  "hotelname": "Iberostar Playa de Muro",
  "hotelstars": 4.0
}
```

### Offer
```json
{
  "hotelid": 90,
  "outbounddeparturedatetime": "2022-10-05T09:30:00+02:00",
  "inbounddeparturedatetime": "2022-10-12T08:35:00+02:00",
  "countadults": 1,
  "countchildren": 1,
  "price": 1243,
  "inbounddepartureairport": "PMI",
  "inboundarrivalairport": "DUS",
  "inboundarrivaldatetime": "2022-10-12T14:40:00+02:00",
  "outbounddepartureairport": "DUS",
  "outboundarrivalairport": "PMI",
  "outboundarrivaldatetime": "2022-10-05T14:25:00+02:00",
  "mealtype": "halfboard",
  "oceanview": false,
  "roomtype": "double"
}
```
