package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(
	db *pgxpool.Pool,
) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Create(
	ctx context.Context,
	userID string,
	tokenHash []byte,
	expiresAt time.Time,
	userAgent *string,
	ipAddress *string,
) (*domain.Session, error) {
	const query = `
		INSERT INTO sessions (
			user_id,
			token_hash,
			expires_at,
			user_agent,
			ip_address
		)
		VALUES (
			CAST($1 AS uuid),
			$2,
			$3,
			$4,
			CAST($5 AS inet)
		)
		RETURNING
			CAST(id AS text),
			CAST(user_id AS text),
			token_hash,
			expires_at,
			created_at,
			revoked_at,
			user_agent,
			CAST(ip_address AS text)
	`

	var session domain.Session

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		tokenHash,
		expiresAt,
		userAgent,
		ipAddress,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.RevokedAt,
		&session.UserAgent,
		&session.IPAddress,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"create session: %w",
			err,
		)
	}

	return &session, nil
}

func (r *SessionRepository) GetActiveByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) (*domain.Session, error) {
	const query = `
		SELECT
			CAST(id AS text),
			CAST(user_id AS text),
			token_hash,
			expires_at,
			created_at,
			revoked_at,
			user_agent,
			CAST(ip_address AS text)
		FROM sessions
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`

	var session domain.Session

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.RevokedAt,
		&session.UserAgent,
		&session.IPAddress,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"get active session: %w",
			err,
		)
	}

	return &session, nil
}

func (r *SessionRepository) RevokeByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) error {
	const query = `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
	`

	if _, err := r.db.Exec(
		ctx,
		query,
		tokenHash,
	); err != nil {
		return fmt.Errorf(
			"revoke session: %w",
			err,
		)
	}

	return nil
}
