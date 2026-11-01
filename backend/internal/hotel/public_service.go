package hotel

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

var ErrInvalidPublicHotelQuery = errors.New(
	"invalid public hotel query",
)

type PublicHotelRepository interface {
	ListPublicByCity(
		ctx context.Context,
		city string,
		limit int,
		offset int,
	) ([]domain.Hotel, error)
}

type PublicService struct {
	repository PublicHotelRepository
}

func NewPublicService(
	repository PublicHotelRepository,
) *PublicService {
	return &PublicService{
		repository: repository,
	}
}

func (s *PublicService) ListByCity(
	ctx context.Context,
	city string,
	limit int,
	offset int,
) ([]domain.Hotel, error) {
	city = strings.TrimSpace(city)

	cityLength := utf8.RuneCountInString(city)

	if cityLength < 2 || cityLength > 120 {
		return nil, ErrInvalidPublicHotelQuery
	}

	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		return nil, ErrInvalidPublicHotelQuery
	}

	return s.repository.ListPublicByCity(
		ctx,
		city,
		limit,
		offset,
	)
}
