ALTER TABLE users
    ADD COLUMN phone_encrypted BYTEA,
    ADD COLUMN phone_lookup_hash BYTEA;

ALTER TABLE phone_verifications
    ADD COLUMN phone_lookup_hash BYTEA;

ALTER TABLE users
    ADD CONSTRAINT users_phone_lookup_hash_length_check
        CHECK (
            phone_lookup_hash IS NULL
            OR octet_length(phone_lookup_hash) = 32
        );

ALTER TABLE users
    ADD CONSTRAINT users_phone_encrypted_not_empty_check
        CHECK (
            phone_encrypted IS NULL
            OR octet_length(phone_encrypted) > 0
        );

ALTER TABLE phone_verifications
    ADD CONSTRAINT phone_verifications_lookup_hash_length_check
        CHECK (
            phone_lookup_hash IS NULL
            OR octet_length(phone_lookup_hash) = 32
        );

CREATE UNIQUE INDEX users_phone_lookup_hash_unique_idx
    ON users (phone_lookup_hash)
    WHERE phone_lookup_hash IS NOT NULL;

CREATE INDEX phone_verifications_lookup_hash_idx
    ON phone_verifications (
        phone_lookup_hash,
        created_at DESC
    )
    WHERE phone_lookup_hash IS NOT NULL;