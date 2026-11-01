ALTER TABLE hotels
    ADD COLUMN latitude DOUBLE PRECISION,
    ADD COLUMN longitude DOUBLE PRECISION;

ALTER TABLE hotels
    ADD CONSTRAINT hotels_coordinates_pair_check
    CHECK ( 
        (
            latitude IS NULL
            AND longitude IS NULL
        )
        OR 
        (
            latitude IS NOT NULL
            AND longitude IS NOT NULL
        )
    );

ALTER TABLE hotels
    ADD CONSTRAINT hotels_latitude_check
    CHECK (
        latitude IS NULL 
        OR latitude BETWEEN -90 AND 90
    );

ALTER TABLE hotels
    ADD CONSTRAINT hotels_longitude_check
    CHECK (
        longitude IS NULL 
        OR longitude BETWEEN -180 AND 180
    );