CREATE KEYSPACE IF NOT EXISTS holidays WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

-- Hotels table (small)
CREATE TABLE IF NOT EXISTS holidays.hotels (
    hotelid int PRIMARY KEY,
    hotelname text,
    hotelstars float
);

-- Offers base table
-- Partition by hotel to make it easy to get cheapest per hotel; use SAI indexes to filter by other attributes
CREATE TABLE IF NOT EXISTS holidays.offers (
    hotelid int,
    outbounddeparturedatetime timestamp,
    inbounddeparturedatetime timestamp,
    countadults int,
    countchildren int,
    price double,
    inbounddepartureairport text,
    inboundarrivalairport text,
    inboundarrivaldatetime timestamp,
    outbounddepartureairport text,
    outboundarrivalairport text,
    outboundarrivaldatetime timestamp,
    mealtype text,
    oceanview boolean,
    roomtype text,
    PRIMARY KEY ((hotelid), price, outbounddeparturedatetime)
) WITH CLUSTERING ORDER BY (price ASC, outbounddeparturedatetime ASC);

-- Indexes intentionally omitted for Scylla OSS compatibility. Consider denormalized tables or secondary indexes if needed later.
