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

type BookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(
	db *pgxpool.Pool,
) *BookingRepository {
	return &BookingRepository{
		db: db,
	}
}

func (r *BookingRepository) ListByHotelUserID(
	ctx context.Context,
	hotelUserID string,
	limit int,
	offset int,
) ([]domain.Booking, error) {
	const query = `
		SELECT
			CAST(b.id AS text),
			CAST(b.guest_user_id AS text),
			CAST(b.hotel_id AS text),
			b.luggage_count,
			b.starts_at,
			b.ends_at,
			b.status,
			b.created_at,
			b.updated_at
		FROM bookings b
		INNER JOIN hotels h ON h.id = b.hotel_id
		WHERE h.user_id = CAST($1 AS uuid)
		ORDER BY
			CASE WHEN b.ends_at >= NOW() THEN 0 ELSE 1 END,
			b.starts_at ASC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, hotelUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list hotel bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var item domain.Booking
		if err := rows.Scan(
			&item.ID,
			&item.GuestUserID,
			&item.HotelID,
			&item.LuggageCount,
			&item.StartsAt,
			&item.EndsAt,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan hotel booking: %w", err)
		}
		bookings = append(bookings, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hotel bookings: %w", err)
	}
	return bookings, nil
}

func (r *BookingRepository) ListByGuestUserID(
	ctx context.Context,
	guestUserID string,
	limit int,
	offset int,
) ([]domain.Booking, error) {
	const query = `
		SELECT
			CAST(id AS text),
			CAST(guest_user_id AS text),
			CAST(hotel_id AS text),
			luggage_count,
			starts_at,
			ends_at,
			status,
			cancelled_at,
			cancellation_refund_eligible,
			created_at,
			updated_at
		FROM bookings
		WHERE guest_user_id = CAST($1 AS uuid)
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, guestUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list guest bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		var item domain.Booking
		if err := rows.Scan(
			&item.ID,
			&item.GuestUserID,
			&item.HotelID,
			&item.LuggageCount,
			&item.StartsAt,
			&item.EndsAt,
			&item.Status,
			&item.CancelledAt,
			&item.CancellationRefundEligible,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan guest booking: %w", err)
		}
		bookings = append(bookings, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate guest bookings: %w", err)
	}
	return bookings, nil
}

func (r *BookingRepository) CancelForGuestUserID(
	ctx context.Context,
	guestUserID string,
	bookingID string,
) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin booking cancellation: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const lockQuery = `
		SELECT status::text, starts_at
		FROM bookings
		WHERE id = CAST($1 AS uuid)
			AND guest_user_id = CAST($2 AS uuid)
		FOR UPDATE
	`

	var currentStatus string
	var startsAt time.Time
	err = tx.QueryRow(ctx, lockQuery, bookingID, guestUserID).Scan(
		&currentStatus,
		&startsAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, fmt.Errorf("lock booking for cancellation: %w", err)
	}

	if currentStatus != string(domain.BookingStatusConfirmed) {
		return nil, domain.ErrBookingInvalidTransition
	}
	if !startsAt.After(time.Now()) {
		return nil, domain.ErrBookingCancellationTooLate
	}

	const updateQuery = `
		UPDATE bookings
		SET
			status = 'cancelled',
			cancelled_at = NOW(),
			cancellation_refund_eligible = (starts_at >= NOW() + INTERVAL '2 hours'),
			updated_at = NOW()
		WHERE id = CAST($1 AS uuid)
		RETURNING
			CAST(id AS text),
			CAST(guest_user_id AS text),
			CAST(hotel_id AS text),
			luggage_count,
			starts_at,
			ends_at,
			status,
			cancelled_at,
			cancellation_refund_eligible,
			created_at,
			updated_at
	`

	var result domain.Booking
	err = tx.QueryRow(ctx, updateQuery, bookingID).Scan(
		&result.ID,
		&result.GuestUserID,
		&result.HotelID,
		&result.LuggageCount,
		&result.StartsAt,
		&result.EndsAt,
		&result.Status,
		&result.CancelledAt,
		&result.CancellationRefundEligible,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel booking: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit booking cancellation: %w", err)
	}
	return &result, nil
}

func (r *BookingRepository) TransitionStatusForHotelUserID(
	ctx context.Context,
	hotelUserID string,
	bookingID string,
	expectedStatus domain.BookingStatus,
	nextStatus domain.BookingStatus,
) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin booking status transition: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const lockQuery = `
		SELECT b.status::text
		FROM bookings b
		INNER JOIN hotels h ON h.id = b.hotel_id
		WHERE b.id = CAST($1 AS uuid)
			AND h.user_id = CAST($2 AS uuid)
		FOR UPDATE OF b
	`

	var currentStatus string
	err = tx.QueryRow(ctx, lockQuery, bookingID, hotelUserID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, fmt.Errorf("lock booking for status transition: %w", err)
	}

	if currentStatus != string(expectedStatus) {
		return nil, domain.ErrBookingInvalidTransition
	}

	const updateQuery = `
		UPDATE bookings
		SET status = $2::booking_status,
			updated_at = NOW()
		WHERE id = CAST($1 AS uuid)
		RETURNING
			CAST(id AS text),
			CAST(guest_user_id AS text),
			CAST(hotel_id AS text),
			luggage_count,
			starts_at,
			ends_at,
			status,
			created_at,
			updated_at
	`

	var result domain.Booking
	err = tx.QueryRow(
		ctx,
		updateQuery,
		bookingID,
		string(nextStatus),
	).Scan(
		&result.ID,
		&result.GuestUserID,
		&result.HotelID,
		&result.LuggageCount,
		&result.StartsAt,
		&result.EndsAt,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update booking status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit booking status transition: %w", err)
	}
	return &result, nil
}

func (r *BookingRepository) Create(
	ctx context.Context,
	guestUserID string,
	hotelID string,
	luggageCount int,
	startsAt time.Time,
	endsAt time.Time,
) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"begin booking transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const hotelQuery = `
		SELECT
			luggage_capacity,
			is_active,
			accepting_bookings,
			latitude IS NOT NULL
				AND longitude IS NOT NULL
		FROM hotels
		WHERE id = CAST($1 AS uuid)
		FOR UPDATE
	`

	var (
		capacity          int
		isActive          bool
		acceptingBookings bool
		hasCoordinates    bool
	)

	err = tx.QueryRow(
		ctx,
		hotelQuery,
		hotelID,
	).Scan(
		&capacity,
		&isActive,
		&acceptingBookings,
		&hasCoordinates,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHotelUnavailable
		}

		return nil, fmt.Errorf(
			"lock hotel for booking: %w",
			err,
		)
	}

	if !isActive || !acceptingBookings || !hasCoordinates || capacity <= 0 {
		return nil, domain.ErrHotelUnavailable
	}

	const capacityQuery = `
		SELECT COALESCE(SUM(luggage_count), 0)
		FROM bookings
		WHERE hotel_id = CAST($1 AS uuid)
			AND (
				status = 'baggage_received'
				OR (
					status IN ('pending', 'confirmed')
					AND starts_at < $3
					AND ends_at > $2
				)
			)
	`

	var occupiedCapacity int64

	err = tx.QueryRow(
		ctx,
		capacityQuery,
		hotelID,
		startsAt,
		endsAt,
	).Scan(&occupiedCapacity)

	if err != nil {
		return nil, fmt.Errorf(
			"calculate occupied capacity: %w",
			err,
		)
	}

	if occupiedCapacity+int64(luggageCount) > int64(capacity) {
		return nil, domain.ErrBookingCapacityExceeded
	}

	const insertQuery = `
		INSERT INTO bookings (
			guest_user_id,
			hotel_id,
			luggage_count,
			starts_at, 
			ends_at, 
			status
		)	
		VALUES (
			CAST($1 AS uuid),
			CAST($2 AS uuid),
			$3,
			$4,
			$5,
			'confirmed'
		)
		RETURNING
			CAST(id AS text),
			CAST(guest_user_id AS text),
			CAST(hotel_id AS text),
			luggage_count,
			starts_at,
			ends_at,
			status,
			created_at,
			updated_at
	`

	var result domain.Booking

	err = tx.QueryRow(
		ctx,
		insertQuery,
		guestUserID,
		hotelID,
		luggageCount,
		startsAt,
		endsAt,
	).Scan(
		&result.ID,
		&result.GuestUserID,
		&result.HotelID,
		&result.LuggageCount,
		&result.StartsAt,
		&result.EndsAt,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"insert booking: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"commit booking transaction: %w",
			err,
		)
	}

	return &result, nil
}
