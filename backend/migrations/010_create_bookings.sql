CREATE TYPE booking_status AS ENUM (
    'pending',
    'confirmed',
    'cancelled',
    'completed'
);

CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    guest_user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE RESTRICT,

    hotel_id UUID NOT NULL
        REFERENCES hotels(id)
        ON DELETE RESTRICT,
    
    luggage_count INTEGER NOT NULL,

    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,

    status booking_status NOT NULL
        DEFAULT 'pending',

    created_at TIMESTAMPTZ NOT NULL
        DEFAULT NOW(),
    
    updated_at TIMESTAMPTZ NOT NULL
        DEFAULT NOW(),

    CONSTRAINT bookings_luggage_count_check
        CHECK (
            luggage_count >= 1
            AND luggage_count <= 50
        ),

    CONSTRAINT bookings_time_range_check
        CHECK (
            ends_at > starts_at
        )
);

CREATE INDEX bookings_guest_created_idx
    ON bookings (
        guest_user_id,
        created_at DESC
    );

CREATE INDEX bookings_hotel_starts_idx
    ON bookings (
        hotel_id,
        starts_at
    );

CREATE INDEX bookings_hotel_active_idx
    ON bookings (
        hotel_id,
        starts_at,
        ends_at
    )
    WHERE status IN (
        'pending',
        'confirmed'
    );
