package hotel

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

var ErrInvalidHotelLocation = errors.New(
	"invalid hotel location",
)

type HotelLocationRepository interface {
	UpdateCoordinatesByUserID(
		ctx context.Context,
		userID string,
		latitude float64,
		longitude float64,
	) (*domain.Hotel, error)
}

type LocationService struct {
	repository HotelLocationRepository
}

func NewLocationService(
	repository HotelLocationRepository,
) *LocationService {
	return &LocationService{
		repository: repository,
	}
}

func (s *LocationService) UpdateCoordinates(
	ctx context.Context,
	userID string,
	latitude float64,
	longitude float64,
) (*domain.Hotel, error) {
	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, ErrInvalidHotelLocation
	}

	if math.IsNaN(latitude) ||
		math.IsNaN(longitude) ||
		latitude < -90 || latitude > 90 ||
		longitude < -180 || longitude > 180 {
		return nil, ErrInvalidHotelLocation
	}

	if math.IsNaN(longitude) ||
		math.IsInf(longitude, 0) ||
		longitude < -180 ||
		longitude > 180 {
		return nil, ErrInvalidHotelLocation
	}

	return s.repository.UpdateCoordinatesByUserID(
		ctx,
		userID,
		latitude,
		longitude,
	)
}
