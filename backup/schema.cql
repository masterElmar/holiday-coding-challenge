
CREATE KEYSPACE holidays WITH replication = {'class': 'org.apache.cassandra.locator.SimpleStrategy', 'replication_factor': '1'} AND durable_writes = true AND tablets = {'enabled': false};

CREATE TABLE holidays.hotels (
    hotelid int,
    hotelname text,
    hotelstars float,
    PRIMARY KEY (hotelid)
) WITH bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE'
    AND tombstone_gc = {'mode': 'timeout', 'propagation_delay_in_seconds': '3600'};


CREATE TABLE holidays.offers (
    hotelid int,
    price double,
    outbounddeparturedatetime timestamp,
    countadults int,
    countchildren int,
    inboundarrivalairport text,
    inboundarrivaldatetime timestamp,
    inbounddepartureairport text,
    inbounddeparturedatetime timestamp,
    mealtype text,
    oceanview boolean,
    outboundarrivalairport text,
    outboundarrivaldatetime timestamp,
    outbounddepartureairport text,
    roomtype text,
    PRIMARY KEY (hotelid, price, outbounddeparturedatetime)
) WITH CLUSTERING ORDER BY (price ASC, outbounddeparturedatetime ASC)
    AND bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE'
    AND tombstone_gc = {'mode': 'timeout', 'propagation_delay_in_seconds': '3600'};

