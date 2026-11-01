package repository

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const otpCooldown = 60 * time.Second

type PhoneVerificationRepository struct {
	db     *pgxpool.Pool
	phones PhoneProtector
}

func NewPhoneVerificationRepository(
	db *pgxpool.Pool,
	phones PhoneProtector,
) *PhoneVerificationRepository {
	return &PhoneVerificationRepository{
		db:     db,
		phones: phones,
	}
}

func (r *PhoneVerificationRepository) Create(
	ctx context.Context,
	phone string,
	codeHash []byte,
	now time.Time,
	expiresAt time.Time,
) (*domain.PhoneVerification, error) {
	lookupHash := r.phones.LookupHash(phone)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"begin create OTP transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := lockPhone(
		ctx,
		tx,
		lookupHash,
	); err != nil {
		return nil, err
	}

	const latestQuery = `
		SELECT created_at
		FROM phone_verifications
		WHERE phone_lookup_hash = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var latestCreatedAt time.Time

	err = tx.QueryRow(
		ctx,
		latestQuery,
		lookupHash,
	).Scan(&latestCreatedAt)

	if err != nil &&
		!errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf(
			"get latest OTP: %w",
			err,
		)
	}

	if err == nil &&
		latestCreatedAt.Add(otpCooldown).After(now) {
		return nil, domain.ErrOTPCooldown
	}

	const invalidateQuery = `
		UPDATE phone_verifications
		SET consumed_at = $2
		WHERE phone_lookup_hash = $1
		  AND consumed_at IS NULL
	`

	if _, err := tx.Exec(
		ctx,
		invalidateQuery,
		lookupHash,
		now,
	); err != nil {
		return nil, fmt.Errorf(
			"invalidate previous OTP: %w",
			err,
		)
	}

	const insertQuery = `
		INSERT INTO phone_verifications (
			phone_lookup_hash,
			code_hash,
			attempts,
			expires_at,
			created_at
		)
		VALUES (
			$1,
			$2,
			0,
			$3,
			$4
		)
		RETURNING
			CAST(id AS text),
			code_hash,
			attempts,
			expires_at,
			consumed_at,
			created_at
	`

	var verification domain.PhoneVerification

	if err := tx.QueryRow(
		ctx,
		insertQuery,
		lookupHash,
		codeHash,
		expiresAt,
		now,
	).Scan(
		&verification.ID,
		&verification.CodeHash,
		&verification.Attempts,
		&verification.ExpiresAt,
		&verification.ConsumedAt,
		&verification.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf(
			"create OTP verification: %w",
			err,
		)
	}

	// Номер существует только в памяти приложения.
	// В phone_verifications он больше не записывается.
	verification.Phone = phone

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"commit create OTP: %w",
			err,
		)
	}

	return &verification, nil
}

func (r *PhoneVerificationRepository) VerifyAndConsume(
	ctx context.Context,
	phone string,
	candidateHash []byte,
	now time.Time,
) (*domain.PhoneVerification, error) {
	lookupHash := r.phones.LookupHash(phone)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"begin verify OTP transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := lockPhone(
		ctx,
		tx,
		lookupHash,
	); err != nil {
		return nil, err
	}

	const query = `
		SELECT
			CAST(id AS text),
			code_hash,
			attempts,
			expires_at,
			consumed_at,
			created_at
		FROM phone_verifications
		WHERE phone_lookup_hash = $1
		  AND consumed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`

	var verification domain.PhoneVerification

	err = tx.QueryRow(
		ctx,
		query,
		lookupHash,
	).Scan(
		&verification.ID,
		&verification.CodeHash,
		&verification.Attempts,
		&verification.ExpiresAt,
		&verification.ConsumedAt,
		&verification.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOTPNotFound
		}

		return nil, fmt.Errorf(
			"get OTP verification: %w",
			err,
		)
	}

	verification.Phone = phone

	if !verification.ExpiresAt.After(now) {
		return nil, domain.ErrOTPExpired
	}

	if verification.Attempts >= 5 {
		return nil, domain.ErrOTPTooManyAttempts
	}

	if !hmac.Equal(
		verification.CodeHash,
		candidateHash,
	) {
		newAttempts := verification.Attempts + 1

		const updateAttemptsQuery = `
			UPDATE phone_verifications
			SET attempts = $2
			WHERE id = $1
		`

		if _, err := tx.Exec(
			ctx,
			updateAttemptsQuery,
			verification.ID,
			newAttempts,
		); err != nil {
			return nil, fmt.Errorf(
				"increment OTP attempts: %w",
				err,
			)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf(
				"commit invalid OTP attempt: %w",
				err,
			)
		}

		if newAttempts >= 5 {
			return nil, domain.ErrOTPTooManyAttempts
		}

		return nil, domain.ErrOTPInvalid
	}

	const consumeQuery = `
		UPDATE phone_verifications
		SET consumed_at = $2
		WHERE id = $1
		  AND consumed_at IS NULL
	`

	if _, err := tx.Exec(
		ctx,
		consumeQuery,
		verification.ID,
		now,
	); err != nil {
		return nil, fmt.Errorf(
			"consume OTP: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"commit OTP verification: %w",
			err,
		)
	}

	verification.ConsumedAt = &now

	return &verification, nil
}

func lockPhone(
	ctx context.Context,
	tx pgx.Tx,
	lookupHash []byte,
) error {
	lockKey := phoneLockKey(lookupHash)

	const query = `
		SELECT pg_advisory_xact_lock($1)
	`

	if _, err := tx.Exec(
		ctx,
		query,
		lockKey,
	); err != nil {
		return fmt.Errorf(
			"lock phone verification: %w",
			err,
		)
	}

	return nil
}

func phoneLockKey(
	lookupHash []byte,
) int64 {
	hash := sha256.Sum256(lookupHash)

	return int64(
		binary.BigEndian.Uint64(hash[:8]),
	)
}
