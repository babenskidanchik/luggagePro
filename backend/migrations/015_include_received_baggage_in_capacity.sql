DROP INDEX IF EXISTS bookings_hotel_active_idx;

CREATE INDEX bookings_hotel_active_idx
    ON bookings (
        hotel_id,
        starts_at,
        ends_at
    )
    WHERE status IN (
        'pending',
        'confirmed',
        'baggage_received'
    );
