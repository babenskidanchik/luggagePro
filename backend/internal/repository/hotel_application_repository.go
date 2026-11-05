package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HotelApplicationRepository struct {
	db *pgxpool.Pool
}

func NewHotelApplicationRepository(
	db *pgxpool.Pool,
) *HotelApplicationRepository {
	return &HotelApplicationRepository{
		db: db,
	}
}

func (r *HotelApplicationRepository) Create(
	ctx context.Context,
	application *domain.HotelApplication,
) (*domain.HotelApplication, error) {
	const query = `
		INSERT INTO hotel_applications (
			user_id,
			legal_name,
			display_name,
			tax_id,
			country_code,
			city,
			address
		)
		VALUES (
			CAST($1 AS uuid),
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		RETURNING
			CAST(id AS text),
			CAST(user_id AS text),
			legal_name,
			display_name,
			tax_id,
			country_code,
			city,
			address,
			status,
			rejection_reason,
			CAST(reviewed_by AS text),
			reviewed_at,
			created_at,
			updated_at
	`

	var created domain.HotelApplication

	err := r.db.QueryRow(
		ctx,
		query,
		application.UserID,
		application.LegalName,
		application.DisplayName,
		application.TaxID,
		application.CountryCode,
		application.City,
		application.Address,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.LegalName,
		&created.DisplayName,
		&created.TaxID,
		&created.CountryCode,
		&created.City,
		&created.Address,
		&created.Status,
		&created.RejectionReason,
		&created.ReviewedBy,
		&created.ReviewedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "hotel_applications_user_unique":
				return nil, domain.ErrHotelApplicationAlreadyExists
			case "hotel_application_active_tax_id_unique_idx":
				return nil, domain.ErrHotelTaxIDAlreadyInUse
			}
		}

		return nil, fmt.Errorf(
			"create hotel application: %w",
			err,
		)
	}

	return &created, nil
}

func (r *HotelApplicationRepository) ListPending(
	ctx context.Context,
	limit int,
	offset int,
) ([]domain.HotelApplication, error) {
	const query = `
		SELECT
			CAST(id AS text),
			CAST(user_id AS text),
			legal_name,
			display_name,
			tax_id,
			country_code,
			city,
			address,
			status,
			rejection_reason,
			CAST(reviewed_by AS text),
			reviewed_at,
			created_at,
			updated_at
		FROM hotel_applications
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
		OFFSET $2
	`

	rows, err := r.db.Query(
		ctx,
		query,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list pending hotel applications: %w",
			err,
		)
	}
	defer rows.Close()

	applications := make(
		[]domain.HotelApplication,
		0,
	)

	for rows.Next() {
		var application domain.HotelApplication

		if err := rows.Scan(
			&application.ID,
			&application.UserID,
			&application.LegalName,
			&application.DisplayName,
			&application.TaxID,
			&application.CountryCode,
			&application.City,
			&application.Address,
			&application.Status,
			&application.RejectionReason,
			&application.ReviewedBy,
			&application.ReviewedAt,
			&application.CreatedAt,
			&application.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan pending hotel application: %w",
				err,
			)
		}

		applications = append(applications, application)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate pending hotel applications: %w",
			err,
		)
	}

	return applications, nil
}

func (r *HotelApplicationRepository) Approve(
	ctx context.Context,
	applicationID string,
	moderatorID string,
) (*domain.Hotel, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"begin approve hotel application: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const applicationQuery = `
		SELECT
			CAST(user_id AS text),
			display_name,
			country_code,
			city, 
			address,
			status
		FROM hotel_applications
		WHERE id = CAST($1 AS uuid)
		FOR UPDATE
	`

	var (
		userID      string
		displayName string
		countryCode string
		city        string
		address     string
		status      domain.HotelApplicationStatus
	)

	err = tx.QueryRow(
		ctx,
		applicationQuery,
		applicationID,
	).Scan(
		&userID,
		&displayName,
		&countryCode,
		&city,
		&address,
		&status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil,
				domain.ErrHotelApplicationNotFound
		}

		return nil, fmt.Errorf(
			"get hotel application for approval: %w",
			err,
		)
	}

	if status != domain.HotelApplicationPending {
		return nil,
			domain.ErrHotelApplicationNotPending
	}

	const updateApplicationQuery = `
		UPDATE hotel_applications
		SET
			status = 'approved',
			rejection_reason = NULL,
			reviewed_by = CAST($2 AS uuid),
			reviewed_at = NOW(),
			updated_at = NOW()
		WHERE id = CAST($1 AS uuid)
	`

	if _, err := tx.Exec(
		ctx,
		updateApplicationQuery,
		applicationID,
		moderatorID,
	); err != nil {
		return nil, fmt.Errorf(
			"mark hotel application approved: %w",
			err,
		)
	}

	const updateUserQuery = `
		UPDATE users
		SET
			role = 'hotel',
			updated_at = NOW()
		WHERE id = CAST($1 AS uuid)
	`

	result, err := tx.Exec(
		ctx,
		updateUserQuery,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"promote user to hotel: %w",
			err,
		)
	}

	if result.RowsAffected() != 1 {
		return nil, fmt.Errorf(
			"promote user to hotel: expected 1 row, got %d",
			result.RowsAffected(),
		)
	}

	const createHotelQuery = `
		INSERT INTO hotels (
			user_id,
			application_id,
			display_name,
			country_code,
			city,
			address
		)
		VALUES (
			CAST($1 AS uuid),
			CAST($2 AS uuid),
			$3,
			$4,
			$5,
			$6
		)
		RETURNING 
			CAST(id AS text),
			CAST(user_id AS text),
			CAST(application_id AS text),
			display_name,
			country_code,
			city,
			address,
			is_active,
			created_at,
			updated_at
	`

	var hotel domain.Hotel

	err = tx.QueryRow(
		ctx,
		createHotelQuery,
		userID,
		applicationID,
		displayName,
		countryCode,
		city,
		address,
	).Scan(
		&hotel.ID,
		&hotel.UserID,
		&hotel.ApplicationID,
		&hotel.DisplayName,
		&hotel.CountryCode,
		&hotel.City,
		&hotel.Address,
		&hotel.IsActive,
		&hotel.CreatedAt,
		&hotel.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"create hotel: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"commit hotel approval: %w",
			err,
		)
	}

	return &hotel, nil
}
