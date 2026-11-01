CREATE TABLE hotels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL UNIQUE
        REFERENCES users(id)
        ON DELETE CASCADE,

    application_id UUID NOT NULL UNIQUE
        REFERENCES hotel_applications(id)
        ON DELETE RESTRICT,

    display_name VARCHAR(255) NOT NULL,

    country_code CHAR(2) NOT NULL,
    city VARCHAR(120) NOT NULL,
    address VARCHAR(500) NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX hotels_city_active_idx
    ON hotels (country_code, city)
    WHERE is_active = TRUE;