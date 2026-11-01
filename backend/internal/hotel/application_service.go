package hotel

import (
	"context"
	"errors"
	"strings"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

var ErrInvalidApplication = errors.New(
	"invalid hotel application",
)

type ApplicationRepository interface {
	Create(
		ctx context.Context,
		application *domain.HotelApplication,
	) (*domain.HotelApplication, error)
}

type ApplicationInput struct {
	LegalName   string
	DisplayName string
	TaxID       string
	CountryCode string
	City        string
	Address     string
}

type ApplicationService struct {
	repository ApplicationRepository
}

func NewApplicationService(
	repository ApplicationRepository,
) *ApplicationService {
	return &ApplicationService{
		repository: repository,
	}
}

func (s *ApplicationService) Submit(
	ctx context.Context,
	userID string,
	input ApplicationInput,
) (*domain.HotelApplication, error) {
	userID = strings.TrimSpace(userID)

	input.LegalName = strings.TrimSpace(
		input.LegalName,
	)

	input.DisplayName = strings.TrimSpace(
		input.DisplayName,
	)

	input.TaxID = strings.TrimSpace(
		input.TaxID,
	)

	input.CountryCode = strings.ToUpper(
		strings.TrimSpace(input.CountryCode),
	)

	input.City = strings.TrimSpace(
		input.City,
	)

	input.Address = strings.TrimSpace(
		input.Address,
	)

	if !validLength(userID, 1, 100) ||
		!validLength(input.LegalName, 2, 255) ||
		!validLength(input.DisplayName, 2, 255) ||
		!validLength(input.TaxID, 2, 64) ||
		!validCountryCode(input.CountryCode) ||
		!validLength(input.City, 2, 120) ||
		!validLength(input.Address, 5, 500) {
		return nil, ErrInvalidApplication
	}

	application := &domain.HotelApplication{
		UserID:      userID,
		LegalName:   input.LegalName,
		DisplayName: input.DisplayName,
		TaxID:       input.TaxID,
		CountryCode: input.CountryCode,
		City:        input.City,
		Address:     input.Address,
	}

	return s.repository.Create(
		ctx,
		application,
	)
}

func validLength(
	value string,
	minLength int,
	maxLength int,
) bool {
	length := len([]rune(value))

	return length >= minLength &&
		length <= maxLength
}

func validCountryCode(
	code string,
) bool {
	if len(code) != 2 {
		return false
	}

	for _, char := range code {
		if char < 'A' || char > 'Z' {
			return false
		}
	}

	return true
}
