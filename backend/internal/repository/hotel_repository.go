package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HotelRepository struct {
	db *pgxpool.Pool
}

func NewHotelRepository(
	db *pgxpool.Pool,
) *HotelRepository {
	return &HotelRepository{
		db: db,
	}
}

func (r *HotelRepository) UpdateCoordinatesByUserID(
	ctx context.Context,
	userID string,
	latitude float64,
	longitude float64,
) (*domain.Hotel, error) {
	const query = `
		UPDATE hotels
		SET
			latitude = $2,
			longitude = $3,
			updated_at = NOW()
		WHERE user_id = CAST($1 AS uuid)
		RETURNING
			CAST(id AS text),
			CAST(user_id AS text),
			CAST(application_id AS text),
			venue_type,
			display_name,
			country_code,
			city,
			address,
			latitude,
			longitude,
			luggage_capacity,
			is_active,
			created_at,
			updated_at
	`

	var hotel domain.Hotel

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		latitude,
		longitude,
	).Scan(
		&hotel.ID,
		&hotel.UserID,
		&hotel.ApplicationID,
		&hotel.VenueType,
		&hotel.DisplayName,
		&hotel.CountryCode,
		&hotel.City,
		&hotel.Address,
		&hotel.Latitude,
		&hotel.Longitude,
		&hotel.LuggageCapacity,
		&hotel.IsActive,
		&hotel.CreatedAt,
		&hotel.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHotelNotFound
		}

		return nil, fmt.Errorf(
			"update hotel coordinates: %w",
			err,
		)
	}

	return &hotel, nil
}

func (r *HotelRepository) UpdateBookingSettingsByUserID(
	ctx context.Context,
	userID string,
	luggageCapacity int,
	acceptingBookings bool,
) (*domain.Hotel, error) {
	const query = `
		UPDATE hotels
		SET
			luggage_capacity = $2,
			accepting_bookings = $3,
			updated_at = NOW()
		WHERE user_id = CAST($1 AS uuid)
		RETURNING luggage_capacity, accepting_bookings
	`

	var result domain.Hotel
	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		luggageCapacity,
		acceptingBookings,
	).Scan(
		&result.LuggageCapacity,
		&result.AcceptingBookings,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHotelNotFound
		}
		return nil, fmt.Errorf("update hotel booking settings: %w", err)
	}

	return &result, nil
}

func (r *HotelRepository) ListPublicByCity(
	ctx context.Context,
	city string,
	limit int,
	offset int,
) ([]domain.Hotel, error) {
	const query = `
		SELECT 
			CAST(id AS text),
			CAST(user_id AS text),
			CAST(application_id AS text),
			display_name,
			country_code,
			city,
			address,
			latitude,
			longitude,
			venue_type,
			luggage_capacity,
			is_active,
			created_at,
			updated_at
		FROM hotels
		WHERE
		  	is_active = TRUE
			AND latitude IS NOT NULL
			AND longitude IS NOT NULL
			AND city = $1
		ORDER BY display_name ASC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		city,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list public hotels by city: %w",
			err,
		)
	}
	defer rows.Close()

	hotels := make([]domain.Hotel, 0)

	for rows.Next() {
		var hotel domain.Hotel

		if err := rows.Scan(
			&hotel.ID,
			&hotel.UserID,
			&hotel.ApplicationID,
			&hotel.DisplayName,
			&hotel.CountryCode,
			&hotel.City,
			&hotel.Address,
			&hotel.Latitude,
			&hotel.Longitude,
			&hotel.VenueType,
			&hotel.LuggageCapacity,
			&hotel.IsActive,
			&hotel.CreatedAt,
			&hotel.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan hotel: %w",
				err,
			)
		}

		hotels = append(
			hotels,
			hotel,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate hotels: %w",
			err,
		)
	}
	return hotels, nil
}
