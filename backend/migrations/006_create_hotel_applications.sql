CREATE TYPE hotel_application_status AS ENUM (
    'pending',
    'approved',
    'rejected'
);

CREATE TABLE hotel_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    legal_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,

    tax_id VARCHAR(64) NOT NULL,

    country_code CHAR(2) NOT NULL,
    city VARCHAR(120) NOT NULL,
    address VARCHAR(500) NOT NULL,

    status hotel_application_status
        NOT NULL
        DEFAULT 'pending',

    rejection_reason TEXT,

    reviewed_by UUID
        REFERENCES users(id)
        ON DELETE SET NULL,

    reviewed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ
        NOT NULL
        DEFAULT NOW(),

    updated_at TIMESTAMPTZ
        NOT NULL
        DEFAULT NOW(),

    CONSTRAINT hotel_applications_user_unique
        UNIQUE (user_id),

    CONSTRAINT hotel_application_review_check
        CHECK (
            (
                status = 'pending'
                AND reviewed_by IS NULL
                AND reviewed_at IS NULL
            )
            OR
            (
                status IN ('approved', 'rejected')
                AND reviewed_by IS NOT NULL
                AND reviewed_at IS NOT NULL
            )
        ),

    CONSTRAINT hotel_application_rejection_reason_check
        CHECK (
            status <> 'rejected'
            OR rejection_reason IS NOT NULL
        )
);

CREATE UNIQUE INDEX hotel_applications_active_tax_id_unique_idx
    ON hotel_applications (
        country_code,
        tax_id
    )
    WHERE status IN ('pending', 'approved');

CREATE INDEX hotel_applications_status_created_idx
    ON hotel_applications (
        status,
        created_at
    );