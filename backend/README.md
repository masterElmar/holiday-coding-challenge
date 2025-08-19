# Holiday Challenge Backend

## Umgebungsvariablen

Erstelle optional eine `.env` oder setze Umgebungsvariablen. Der Backend-Server nutzt ScyllaDB (gocql):

```env
# Server
PORT=8090

# Datenquellen (CSV Fallback/Import)
HOTELS_DATA_PATH=../data/hotels.csv
OFFERS_DATA_PATH=../data/offers.csv

# Scylla
SCYLLA_HOSTS=127.0.0.1
SCYLLA_PORT=9042
SCYLLA_KEYSPACE=holidays
# optional
# SCYLLA_USERNAME=
# SCYLLA_PASSWORD=
# SCYLLA_CONSISTENCY=QUORUM
# SCYLLA_LOCAL_DC=
```

## Verfügbare Umgebungsvariablen

| Variable | Beschreibung | Standard |
|----------|--------------|----------|
| `PORT` | Server Port | `8090` |
| `HOTELS_DATA_PATH` | Pfad Hotels CSV (für Initialimport) | `../data/hotels.csv` |
| `OFFERS_DATA_PATH` | Pfad Offers CSV (für Import-Tool) | `../data/offers.csv` |
| `SCYLLA_HOSTS` | Kommagetrennte Hosts | `127.0.0.1` |
| `SCYLLA_PORT` | Port | `9042` |
| `SCYLLA_KEYSPACE` | Keyspace | `holidays` |

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

Optional: Großes Offers-CSV nach Scylla importieren:

```bash
cd backend
go run cmd/import-offers/main.go -offers ../data/offers.csv
```

## Verfügbare Endpunkte

- `GET /api/health` - Gesundheitsstatus
- `GET /api/stats` - Statistiken
- `GET /bestOffersByHotel` - Beste (günstigste) Angebote je Hotel nach Suche
- `GET /hotels/{id}/offers` - Alle Angebote für ein Hotel

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
