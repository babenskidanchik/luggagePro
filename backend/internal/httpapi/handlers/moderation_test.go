package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type fakeModerationService struct {
	approveResult *domain.Hotel
	approveErr    error
	applicationID string
	moderatorID   string
	limit         int
	offset        int
}

func (s *fakeModerationService) ListPending(
	_ context.Context,
	limit int,
	offset int,
) ([]domain.HotelApplication, error) {
	s.limit = limit
	s.offset = offset

	return []domain.HotelApplication{
		{
			ID:          "application-id",
			DisplayName: "Batumi Luggage Spot",
			Status:      domain.HotelApplicationPending,
		},
	}, nil
}

func TestModerationHandlerListPending(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeModerationService{}

	handler := NewModerationHandler(
		service,
	)

	router := gin.New()

	router.GET(
		"/test",
		handler.ListPending,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test?limit=20&offset=5",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if service.limit != 20 {
		t.Fatalf(
			"expected limit 20, got %d",
			service.limit,
		)
	}

	if service.offset != 5 {
		t.Fatalf(
			"expected offset 5, got %d",
			service.offset,
		)
	}
}

func TestModerationHandlerRejectsInvalidLimit(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeModerationService{}

	handler := NewModerationHandler(
		service,
	)

	router := gin.New()

	router.GET(
		"/test",
		handler.ListPending,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test?limit=abc",
		nil,
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

func (s *fakeModerationService) Approve(
	ctx context.Context,
	applicationID string,
	moderatorID string,
) (*domain.Hotel, error) {
	s.applicationID = applicationID
	s.moderatorID = moderatorID

	if s.approveErr != nil {
		return nil, s.approveErr
	}

	return s.approveResult, nil
}

func TestModerationHandlerApprove(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeModerationService{
		approveResult: &domain.Hotel{
			ID:          "hotel-id",
			DisplayName: "Batumi Luggage Spot",
			IsActive:    true,
		},
	}

	handler := NewModerationHandler(
		service,
	)

	router := gin.New()

	router.POST(
		"/test/:id",

		func(c *gin.Context) {
			c.Set(
				middleware.CurrentUserContextKey,
				&domain.User{
					ID:   "moderator-id",
					Role: domain.RoleModerator,
				},
			)
			c.Next()
		},

		handler.Approve,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/test/application-id",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if service.applicationID != "application-id" {
		t.Fatalf(
			"unexpected application id: %q",
			service.applicationID,
		)
	}

	if service.moderatorID != "moderator-id" {
		t.Fatalf(
			"unexpected moderator id: %q",
			service.moderatorID,
		)
	}
}

func TestModerationHandlerApproveRejectsNonPending(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeModerationService{
		approveErr: domain.ErrHotelApplicationNotPending,
	}

	handler := NewModerationHandler(
		service,
	)

	router := gin.New()

	router.POST(
		"/test/:id",

		func(c *gin.Context) {
			c.Set(
				middleware.CurrentUserContextKey,
				&domain.User{
					ID:   "moderator-id",
					Role: domain.RoleModerator,
				},
			)
			c.Next()
		},

		handler.Approve,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/test/application-id",
		nil,
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
