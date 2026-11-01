package hotel

import (
	"context"
	"errors"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type fakeHotelLocationRepository struct {
	userID    string
	latitude  float64
	longitude float64
	result    *domain.Hotel
	err       error
}

func (r *fakeHotelLocationRepository) UpdateCoordinatesByUserID(
	_ context.Context,
	userID string,
	latitude float64,
	longitude float64,
) (*domain.Hotel, error) {
	r.userID = userID
	r.latitude = latitude
	r.longitude = longitude

	return r.result, r.err
}

func TestLocationServiceUpdateCoordinates(t *testing.T) {
	latitude := 41.6461
	longitude := 41.6405

	repository := &fakeHotelLocationRepository{
		result: &domain.Hotel{
			ID:        "hotel-id",
			UserID:    "user-id",
			Latitude:  &latitude,
			Longitude: &longitude,
		},
	}

	service := NewLocationService(repository)

	result, err := service.UpdateCoordinates(
		context.Background(),
		"user-id",
		latitude,
		longitude,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected hotel")
	}

	if repository.userID != "user-id" {
		t.Fatalf(
			"unexpected user id: %q",
			repository.userID,
		)
	}

	if repository.latitude != latitude {
		t.Fatalf(
			"unexpected latitude: %f",
			repository.latitude,
		)
	}

	if repository.longitude != longitude {
		t.Fatalf(
			"unexpected longitude: %f",
			repository.longitude,
		)
	}
}

func TestLocationServiceRejectsInvalidCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{
			name:      "latitude too low",
			latitude:  -90.1,
			longitude: 41,
		},
		{
			name:      "latitude too high",
			latitude:  90.1,
			longitude: 41,
		},
		{
			name:      "longitude too low",
			latitude:  41,
			longitude: -180.1,
		},
		{
			name:      "longitude too high",
			latitude:  41,
			longitude: 180.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakeHotelLocationRepository{}
			service := NewLocationService(repository)

			_, err := service.UpdateCoordinates(
				context.Background(),
				"user-id",
				tt.latitude,
				tt.longitude,
			)

			if !errors.Is(
				err,
				ErrInvalidHotelLocation,
			) {
				t.Fatalf(
					"expected ErrInvalidHotelLocation, got %v",
					err,
				)
			}

			if repository.userID != "" {
				t.Fatal(
					"repository must not be called",
				)
			}
		})
	}
}
