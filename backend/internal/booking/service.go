package booking

import (
	"context"
	"strings"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type BookingRepository interface {
	Create(
		ctx context.Context,
		guestUserID string,
		hotelID string,
		luggageCount int,
		startsAt time.Time,
		endsAt time.Time,
	) (*domain.Booking, error)
}

type Service struct {
	repository BookingRepository
}

func NewService(repository BookingRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	ctx context.Context,
	guestUserID string,
	hotelID string,
	luggageCount int,
	startsAt time.Time,
	endsAt time.Time,
) (*domain.Booking, error) {
	guestUserID = strings.TrimSpace(guestUserID)
	hotelID = strings.TrimSpace(hotelID)

	if guestUserID == "" ||
		hotelID == "" ||
		luggageCount < 1 ||
		luggageCount > 50 ||
		startsAt.IsZero() ||
		endsAt.IsZero() ||
		!endsAt.After(startsAt) {
		return nil, domain.ErrInvalidBooking
	}

	return s.repository.Create(
		ctx,
		guestUserID,
		hotelID,
		luggageCount,
		startsAt,
		endsAt,
	)
}
