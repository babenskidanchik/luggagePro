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

func (r *BookingRepository) Create(
	ctx context.Context,
	guestUserID string,
	hotelID string,
	luggageCount int,
	startsAt time.Time,
	endsAt time.Time,
) (*domain.Booking, error) {
	const query = `
		INSERT INTO bookings (
			guest_user_id,
			hotel_id,
			luggage_count,
			starts_at,
			ends_at
		)
		SELECT
			CAST($1 AS uuid),
			h.id,
			$3,
			$4,
			$5
		FROM hotels h
		WHERE
			h.id = CAST($2 AS uuid)
			AND h.is_active = TRUE
			AND h.latitude IS NOT NULL
			AND h.longitude IS NOT NULL
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

	var booking domain.Booking

	err := r.db.QueryRow(
		ctx,
		query,
		guestUserID,
		hotelID,
		luggageCount,
		startsAt,
		endsAt,
	).Scan(
		&booking.ID,
		&booking.GuestUserID,
		&booking.HotelID,
		&booking.LuggageCount,
		&booking.StartsAt,
		&booking.EndsAt,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHotelUnavailable
		}

		return nil, fmt.Errorf(
			"create booking: %w",
			err,
		)
	}

	return &booking, nil
}
