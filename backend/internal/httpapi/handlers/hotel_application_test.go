package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/hotel"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type fakeHotelApplicationService struct {
	result *domain.HotelApplication
	err    error
	userID string
}

func (s *fakeHotelApplicationService) Submit(
	_ context.Context,
	userID string,
	_ hotel.ApplicationInput,
) (*domain.HotelApplication, error) {
	s.userID = userID

	if s.err != nil {
		return nil, s.err
	}

	return s.result, nil
}

func TestHotelApplicationHandlerSubmit(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHotelApplicationService{
		result: &domain.HotelApplication{
			ID:          "application-id",
			UserID:      "user-id",
			LegalName:   "Example LLC",
			DisplayName: "Example Hotel",
			TaxID:       "123456",
			CountryCode: "GE",
			City:        "Batumi",
			Address:     "1 Rustaveli Street",
			Status:      domain.HotelApplicationPending,
		},
	}

	handler := NewHotelApplicationHandler(
		service,
	)

	router := gin.New()

	router.POST(
		"/test",
		func(c *gin.Context) {
			c.Set(
				middleware.CurrentUserContextKey,
				&domain.User{
					ID:   "user-id",
					Role: domain.RoleGuest,
				},
			)

			c.Next()
		},
		handler.Submit,
	)

	body := `{
		"legal_name":"Example LLC",
		"display_name":"Example Hotel",
		"tax_id":"123456",
		"country_code":"GE",
		"city":"Batumi",
		"address":"1 Rustaveli Street"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		strings.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if service.userID != "user-id" {
		t.Fatalf(
			"expected authenticated user id, got %q",
			service.userID,
		)
	}
}

func TestHotelApplicationHandlerRejectsInvalidApplication(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHotelApplicationService{
		err: hotel.ErrInvalidApplication,
	}

	handler := NewHotelApplicationHandler(
		service,
	)

	router := gin.New()

	router.POST(
		"/test",
		func(c *gin.Context) {
			c.Set(
				middleware.CurrentUserContextKey,
				&domain.User{
					ID:   "user-id",
					Role: domain.RoleGuest,
				},
			)
			c.Next()
		},
		handler.Submit,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		strings.NewReader(`{}`),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400, got %d",
			recorder.Code,
		)
	}
}

func TestHotelApplicationHandlerRejectsConflict(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHotelApplicationService{
		err: domain.ErrHotelApplicationAlreadyExists,
	}

	handler := NewHotelApplicationHandler(
		service,
	)

	router := gin.New()

	router.POST(
		"/test",
		func(c *gin.Context) {
			c.Set(
				middleware.CurrentUserContextKey,
				&domain.User{
					ID:   "user-id",
					Role: domain.RoleGuest,
				},
			)
			c.Next()
		},
		handler.Submit,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		strings.NewReader(`{
			"legal_name":"Example LLC",
			"display_name":"Example Hotel",
			"tax_id":"123456",
			"country_code":"GE",
			"city":"Batumi",
			"address":"1 Rustaveli Street"
		}`),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected 409, got %d",
			recorder.Code,
		)
	}
}
