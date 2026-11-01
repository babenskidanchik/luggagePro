package hotel

import (
	"context"
	"errors"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type fakeApplicationRepository struct {
	created *domain.HotelApplication
	err     error
}

func (r *fakeApplicationRepository) Create(
	_ context.Context,
	application *domain.HotelApplication,
) (*domain.HotelApplication, error) {
	if r.err != nil {
		return nil, r.err
	}

	r.created = application

	result := *application
	result.ID = "application-id"
	result.Status = domain.HotelApplicationPending

	return &result, nil
}

func TestApplicationServiceSubmit(
	t *testing.T,
) {
	repository := &fakeApplicationRepository{}

	service := NewApplicationService(
		repository,
	)

	result, err := service.Submit(
		context.Background(),
		"user-id",
		ApplicationInput{
			LegalName:   "  Example LLC  ",
			DisplayName: "  Example Hotel  ",
			TaxID:       "  123456789  ",
			CountryCode: " ge ",
			City:        " Batumi ",
			Address:     " 1 Rustaveli Street ",
		},
	)
	if err != nil {
		t.Fatalf(
			"submit application: %v",
			err,
		)
	}

	if result.Status != domain.HotelApplicationPending {
		t.Fatalf(
			"expected pending status, got %q",
			result.Status,
		)
	}

	if repository.created.CountryCode != "GE" {
		t.Fatalf(
			"expected GE, got %q",
			repository.created.CountryCode,
		)
	}

	if repository.created.LegalName != "Example LLC" {
		t.Fatalf(
			"unexpected legal name: %q",
			repository.created.LegalName,
		)
	}
}

func TestApplicationServiceRejectsInvalidInput(
	t *testing.T,
) {
	repository := &fakeApplicationRepository{}

	service := NewApplicationService(
		repository,
	)

	_, err := service.Submit(
		context.Background(),
		"user-id",
		ApplicationInput{
			LegalName:   "",
			DisplayName: "Hotel",
			TaxID:       "123",
			CountryCode: "GE",
			City:        "Batumi",
			Address:     "Some address",
		},
	)

	if !errors.Is(
		err,
		ErrInvalidApplication,
	) {
		t.Fatalf(
			"expected invalid application, got %v",
			err,
		)
	}
}

func TestApplicationServiceRejectsBadCountryCode(
	t *testing.T,
) {
	repository := &fakeApplicationRepository{}

	service := NewApplicationService(
		repository,
	)

	_, err := service.Submit(
		context.Background(),
		"user-id",
		ApplicationInput{
			LegalName:   "Example LLC",
			DisplayName: "Example Hotel",
			TaxID:       "123456",
			CountryCode: "G1",
			City:        "Batumi",
			Address:     "Some valid address",
		},
	)

	if !errors.Is(
		err,
		ErrInvalidApplication,
	) {
		t.Fatalf(
			"expected invalid application, got %v",
			err,
		)
	}
}

func TestApplicationServiceReturnsRepositoryError(
	t *testing.T,
) {
	repositoryError := domain.ErrHotelApplicationAlreadyExists

	repository := &fakeApplicationRepository{
		err: repositoryError,
	}

	service := NewApplicationService(
		repository,
	)

	_, err := service.Submit(
		context.Background(),
		"user-id",
		ApplicationInput{
			LegalName:   "Example LLC",
			DisplayName: "Example Hotel",
			TaxID:       "123456",
			CountryCode: "GE",
			City:        "Batumi",
			Address:     "Some valid address",
		},
	)

	if !errors.Is(
		err,
		repositoryError,
	) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}
