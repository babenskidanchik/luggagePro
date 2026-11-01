package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type fakeHotelLocationService struct {
	userID    string
	latitude  float64
	longitude float64
	result    *domain.Hotel
	err       error
}

func (s *fakeHotelLocationService) UpdateCoordinates(
	_ context.Context,
	userID string,
	latitude float64,
	longitude float64,
) (*domain.Hotel, error) {
	s.userID = userID
	s.latitude = latitude
	s.longitude = longitude

	if s.err != nil {
		return nil, s.err
	}

	return s.result, nil
}

func TestHotelLocationHandlerUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	latitude := 41.6461
	longitude := 41.6405

	service := &fakeHotelLocationService{
		result: &domain.Hotel{
			ID:        "hotel-id",
			UserID:    "hotel-user-id",
			Latitude:  &latitude,
			Longitude: &longitude,
		},
	}

	handler := NewHotelLocationHandler(service)

	router := gin.New()

	router.PATCH(
		"/test/location",
		func(c *gin.Context) {
			c.Set(
				middleware.CurrentUserContextKey,
				&domain.User{
					ID:   "hotel-user-id",
					Role: domain.RoleHotel,
				},
			)

			c.Next()
		},
		handler.Update,
	)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/test/location",
		strings.NewReader(
			`{
				"latitude": 41.6461,
				"longitude": 41.6405
			}`,
		),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if service.userID != "hotel-user-id" {
		t.Fatalf(
			"unexpected user id: %q",
			service.userID,
		)
	}

	if service.latitude != latitude {
		t.Fatalf(
			"unexpected latitude: %f",
			service.latitude,
		)
	}

	if service.longitude != longitude {
		t.Fatalf(
			"unexpected longitude: %f",
			service.longitude,
		)
	}
}

func TestHotelLocationHandlerRejectsMissingCoordinates(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHotelLocationService{}

	handler := NewHotelLocationHandler(service)

	router := gin.New()

	router.PATCH(
		"/test/location",
		func(c *gin.Context) {
			c.Set(
				middleware.CurrentUserContextKey,
				&domain.User{
					ID:   "hotel-user-id",
					Role: domain.RoleHotel,
				},
			)

			c.Next()
		},
		handler.Update,
	)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/test/location",
		strings.NewReader(`{}`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if service.userID != "" {
		t.Fatal(
			"service must not be called",
		)
	}
}
