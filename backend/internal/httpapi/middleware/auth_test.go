package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type fakeAuthService struct {
	expectedToken string
	user          *domain.User
	err           error
}

func (f *fakeAuthService) CurrentUser(
	ctx context.Context,
	token string,
) (*domain.User, error) {
	if token != f.expectedToken {
		return nil, domain.ErrSessionNotFound
	}

	if f.err != nil {
		return nil, f.err
	}

	return f.user, nil
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	service := &fakeAuthService{}

	router.GET(
		"/protected",
		RequireAuth(service),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			response.Code,
		)
	}
}

func TestRequireAuthAcceptsBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	service := &fakeAuthService{
		expectedToken: "valid-token",
		user: &domain.User{
			ID:       "user-id",
			Role:     domain.RoleGuest,
			IsActive: true,
		},
	}

	router.GET(
		"/protected",
		RequireAuth(service),
		func(c *gin.Context) {
			user, ok := CurrentUser(c)
			if !ok {
				t.Fatal("expected current user")
			}

			if user.ID != "user-id" {
				t.Fatalf(
					"unexpected user id: %s",
					user.ID,
				)
			}

			c.Status(http.StatusOK)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer valid-token",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			response.Code,
		)
	}
}

func TestRequireAuthRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	service := &fakeAuthService{
		expectedToken: "valid-token",
	}

	router.GET(
		"/protected",
		RequireAuth(service),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer wrong-token",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			response.Code,
		)
	}
}

func TestRequireAuthAcceptsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	service := &fakeAuthService{
		expectedToken: "cookie-token",
		user: &domain.User{
			ID:       "user-id",
			Role:     domain.RoleGuest,
			IsActive: true,
		},
	}

	router.GET(
		"/protected",
		RequireAuth(service),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	request.AddCookie(
		&http.Cookie{
			Name:  SessionCookieName,
			Value: "cookie-token",
		},
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			response.Code,
		)
	}
}
