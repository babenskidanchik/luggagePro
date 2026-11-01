ALTER TABLE users
    DROP CONSTRAINT users_identity_required;

ALTER TABLE users
    ADD CONSTRAINT users_identity_required
        CHECK (
            phone_encrypted IS NOT NULL
            OR email IS NOT NULL
        );
    
ALTER TABLE phone_verifications
    ALTER COLUMN phone DROP NOT NULL;