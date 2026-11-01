CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM (
    'guest',
    'hotel',
    'moderator'
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    phone VARCHAR(32),
    email VARCHAR(255),

    role user_role NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_identity_required
        CHECK (phone IS NOT NULL OR email IS NOT NULL)
);

CREATE UNIQUE INDEX users_phone_unique_idx
    ON users (phone)
    WHERE phone IS NOT NULL;

CREATE UNIQUE INDEX users_email_unique_idx
    ON users (LOWER(email))
    WHERE email IS NOT NULL;