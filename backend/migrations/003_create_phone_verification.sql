CREATE TABLE phone_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    phone VARCHAR(32) NOT NULL,
    code_hash BYTEA NOT NULL,

    attempts SMALLINT NOT NULL DEFAULT 0,

    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT phone_verifications_attempts_check
        CHECK (attempts >= 0 AND attempts <= 5),

    CONSTRAINT phone_verifications_expiry_check
        CHECK (expires_at > created_at)
);

CREATE INDEX phone_verifications_phone_idx
    ON phone_verifications (phone, created_at DESC);

CREATE INDEX phone_verifications_active_idx
    ON phone_verifications (phone, created_at DESC)
    WHERE consumed_at IS NULL;