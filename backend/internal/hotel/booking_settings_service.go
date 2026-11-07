package hotel

import (
	"context"
	"errors"
	"strings"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

var ErrInvalidBookingSettings = errors.New("invalid hotel booking settings")

type BookingSettingsRepository interface {
	UpdateBookingSettingsByUserID(
		ctx context.Context,
		userID string,
		luggageCapacity int,
		acceptingBookings bool,
	) (*domain.Hotel, error)
}

type BookingSettingsService struct {
	repository BookingSettingsRepository
}

func NewBookingSettingsService(repository BookingSettingsRepository) *BookingSettingsService {
	return &BookingSettingsService{repository: repository}
}

func (s *BookingSettingsService) Update(
	ctx context.Context,
	userID string,
	luggageCapacity int,
	acceptingBookings bool,
) (*domain.Hotel, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || luggageCapacity < 0 || luggageCapacity > 10000 {
		return nil, ErrInvalidBookingSettings
	}

	return s.repository.UpdateBookingSettingsByUserID(
		ctx,
		userID,
		luggageCapacity,
		acceptingBookings,
	)
}
