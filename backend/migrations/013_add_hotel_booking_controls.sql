ALTER TABLE hotels
ADD COLUMN accepting_bookings BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX hotels_accepting_bookings_idx
    ON hotels (accepting_bookings)
    WHERE is_active = TRUE;
