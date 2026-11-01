package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type fakePublicHotelService struct {
	city   string
	limit  int
	offset int
	result []domain.Hotel
	err    error
}

func (s *fakePublicHotelService) ListByCity(
	_ context.Context,
	city string,
	limit int,
	offset int,
) ([]domain.Hotel, error) {
	s.city = city
	s.limit = limit
	s.offset = offset

	if s.err != nil {
		return nil, s.err
	}

	return s.result, nil
}

func TestPublicHotelHandlerListByCity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	latitude := 41.6461
	longitude := 41.6405

	service := &fakePublicHotelService{
		result: []domain.Hotel{
			{
				ID:            "hotel-id",
				UserID:        "private-user-id",
				ApplicationID: "private-application-id",
				DisplayName:   "Batumi Luggage Spot",
				CountryCode:   "GE",
				City:          "Batumi",
				Address:       "1 Rustaveli Street",
				Latitude:      &latitude,
				Longitude:     &longitude,
				IsActive:      true,
			},
		},
	}

	handler := NewPublicHotelHandler(service)

	router := gin.New()
	router.GET("/hotels", handler.ListByCity)

	request := httptest.NewRequest(
		http.MethodGet,
		"/hotels?city=Batumi&limit=20&offset=5",
		nil,
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

	if service.city != "Batumi" {
		t.Fatalf(
			"unexpected city: %q",
			service.city,
		)
	}

	if service.limit != 20 {
		t.Fatalf(
			"unexpected limit: %d",
			service.limit,
		)
	}

	if service.offset != 5 {
		t.Fatalf(
			"unexpected offset: %d",
			service.offset,
		)
	}

	body := recorder.Body.String()

	if body == "" {
		t.Fatal("expected response body")
	}
}

func TestPublicHotelHandlerRejectsInvalidLimit(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakePublicHotelService{}
	handler := NewPublicHotelHandler(service)

	router := gin.New()
	router.GET("/hotels", handler.ListByCity)

	request := httptest.NewRequest(
		http.MethodGet,
		"/hotels?city=Batumi&limit=abc",
		nil,
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
}
