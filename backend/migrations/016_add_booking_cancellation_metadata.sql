ALTER TABLE bookings
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN cancellation_refund_eligible BOOLEAN;

UPDATE bookings
SET
    cancelled_at = COALESCE(cancelled_at, updated_at),
    cancellation_refund_eligible = COALESCE(cancellation_refund_eligible, FALSE)
WHERE status = 'cancelled';

ALTER TABLE bookings
    ADD CONSTRAINT bookings_cancellation_metadata_check
    CHECK (
        (status = 'cancelled'
            AND cancelled_at IS NOT NULL
            AND cancellation_refund_eligible IS NOT NULL)
        OR
        (status <> 'cancelled'
            AND cancelled_at IS NULL
            AND cancellation_refund_eligible IS NULL)
    );
