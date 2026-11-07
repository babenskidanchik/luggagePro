package booking

import (
	"context"
	"strings"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type BookingRepository interface {
	ListByGuestUserID(
		ctx context.Context,
		guestUserID string,
		limit int,
		offset int,
	) ([]domain.Booking, error)
	CancelForGuestUserID(
		ctx context.Context,
		guestUserID string,
		bookingID string,
	) (*domain.Booking, error)
	TransitionStatusForHotelUserID(
		ctx context.Context,
		hotelUserID string,
		bookingID string,
		expectedStatus domain.BookingStatus,
		nextStatus domain.BookingStatus,
	) (*domain.Booking, error)
	ListByHotelUserID(
		ctx context.Context,
		hotelUserID string,
		limit int,
		offset int,
	) ([]domain.Booking, error)
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

func (s *Service) ListForGuest(
	ctx context.Context,
	guestUserID string,
	limit int,
	offset int,
) ([]domain.Booking, error) {
	guestUserID = strings.TrimSpace(guestUserID)
	if guestUserID == "" {
		return nil, domain.ErrInvalidBooking
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repository.ListByGuestUserID(
		ctx,
		guestUserID,
		limit,
		offset,
	)
}

func (s *Service) CancelForGuest(
	ctx context.Context,
	guestUserID string,
	bookingID string,
) (*domain.Booking, error) {
	guestUserID = strings.TrimSpace(guestUserID)
	bookingID = strings.TrimSpace(bookingID)
	if guestUserID == "" || bookingID == "" {
		return nil, domain.ErrInvalidBooking
	}
	return s.repository.CancelForGuestUserID(
		ctx,
		guestUserID,
		bookingID,
	)
}

func (s *Service) ListForHotel(
	ctx context.Context,
	hotelUserID string,
	limit int,
	offset int,
) ([]domain.Booking, error) {
	hotelUserID = strings.TrimSpace(hotelUserID)
	if hotelUserID == "" {
		return nil, domain.ErrInvalidBooking
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repository.ListByHotelUserID(
		ctx,
		hotelUserID,
		limit,
		offset,
	)
}

func (s *Service) ConfirmBaggageReceived(
	ctx context.Context,
	hotelUserID string,
	bookingID string,
) (*domain.Booking, error) {
	return s.transitionForHotel(
		ctx,
		hotelUserID,
		bookingID,
		domain.BookingStatusConfirmed,
		domain.BookingStatusBaggageReceived,
	)
}

func (s *Service) Complete(
	ctx context.Context,
	hotelUserID string,
	bookingID string,
) (*domain.Booking, error) {
	return s.transitionForHotel(
		ctx,
		hotelUserID,
		bookingID,
		domain.BookingStatusBaggageReceived,
		domain.BookingStatusCompleted,
	)
}

func (s *Service) transitionForHotel(
	ctx context.Context,
	hotelUserID string,
	bookingID string,
	expectedStatus domain.BookingStatus,
	nextStatus domain.BookingStatus,
) (*domain.Booking, error) {
	hotelUserID = strings.TrimSpace(hotelUserID)
	bookingID = strings.TrimSpace(bookingID)
	if hotelUserID == "" || bookingID == "" {
		return nil, domain.ErrInvalidBooking
	}
	return s.repository.TransitionStatusForHotelUserID(
		ctx,
		hotelUserID,
		bookingID,
		expectedStatus,
		nextStatus,
	)
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
