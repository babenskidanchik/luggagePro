package hotel

import (
	"context"
	"errors"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type fakePublicHotelRepository struct {
	city   string
	limit  int
	offset int
	result []domain.Hotel
	err    error
}

func (r *fakePublicHotelRepository) ListPublicByCity(
	_ context.Context,
	city string,
	limit int,
	offset int,
) ([]domain.Hotel, error) {
	r.city = city
	r.limit = limit
	r.offset = offset

	if r.err != nil {
		return nil, r.err
	}

	return r.result, nil
}

func TestPublicServiceListByCity(t *testing.T) {
	repository := &fakePublicHotelRepository{
		result: []domain.Hotel{
			{
				ID:          "hotel-id",
				DisplayName: "Batumi Luggage Spot",
				City:        "Batumi",
			},
		},
	}

	service := NewPublicService(repository)

	result, err := service.ListByCity(
		context.Background(),
		"  Batumi  ",
		20,
		5,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(result) != 1 {
		t.Fatalf(
			"expected 1 hotel, got %d",
			len(result),
		)
	}

	if repository.city != "Batumi" {
		t.Fatalf(
			"unexpected city: %q",
			repository.city,
		)
	}

	if repository.limit != 20 {
		t.Fatalf(
			"unexpected limit: %d",
			repository.limit,
		)
	}

	if repository.offset != 5 {
		t.Fatalf(
			"unexpected offset: %d",
			repository.offset,
		)
	}
}

func TestPublicServiceListByCityDefaultsLimit(
	t *testing.T,
) {
	repository := &fakePublicHotelRepository{}
	service := NewPublicService(repository)

	_, err := service.ListByCity(
		context.Background(),
		"Batumi",
		0,
		0,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repository.limit != 50 {
		t.Fatalf(
			"expected default limit 50, got %d",
			repository.limit,
		)
	}
}

func TestPublicServiceListByCityCapsLimit(
	t *testing.T,
) {
	repository := &fakePublicHotelRepository{}
	service := NewPublicService(repository)

	_, err := service.ListByCity(
		context.Background(),
		"Batumi",
		500,
		0,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repository.limit != 100 {
		t.Fatalf(
			"expected capped limit 100, got %d",
			repository.limit,
		)
	}
}

func TestPublicServiceRejectsInvalidQuery(
	t *testing.T,
) {
	tests := []struct {
		name   string
		city   string
		offset int
	}{
		{
			name:   "empty city",
			city:   "",
			offset: 0,
		},
		{
			name:   "city too short",
			city:   "A",
			offset: 0,
		},
		{
			name:   "negative offset",
			city:   "Batumi",
			offset: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakePublicHotelRepository{}
			service := NewPublicService(repository)

			_, err := service.ListByCity(
				context.Background(),
				tt.city,
				20,
				tt.offset,
			)

			if !errors.Is(
				err,
				ErrInvalidPublicHotelQuery,
			) {
				t.Fatalf(
					"expected ErrInvalidPublicHotelQuery, got %v",
					err,
				)
			}

			if repository.city != "" {
				t.Fatal(
					"repository must not be called",
				)
			}
		})
	}
}
