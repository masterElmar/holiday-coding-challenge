-- WARNING: the offers dataset is extremely large (~90M rows).
-- Importing via cqlsh COPY can take a very long time. Consider a custom Go ingestor using gocql or scylla-manager loader.

COPY holidays.offers (
  hotelid,
  outbounddeparturedatetime,
  inbounddeparturedatetime,
  countadults,
  countchildren,
  price,
  inbounddepartureairport,
  inboundarrivalairport,
  inboundarrivaldatetime,
  outbounddepartureairport,
  outboundarrivalairport,
  outboundarrivaldatetime,
  mealtype,
  oceanview,
  roomtype
) FROM '/data/offers.csv' WITH HEADER=TRUE AND DELIMITER=','
  AND NULL=''
  AND NUMPROCESSES=4;
