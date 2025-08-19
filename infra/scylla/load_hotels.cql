COPY holidays.hotels (hotelid, hotelname, hotelstars)
FROM '/data/hotels.csv' WITH DELIMITER=';' AND HEADER=TRUE;
