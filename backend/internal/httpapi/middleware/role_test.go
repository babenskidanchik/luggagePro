package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func TestRequireRoleAllowsMatchingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.GET(
		"/test",
		func(c *gin.Context) {
			c.Set(
				CurrentUserContextKey,
				&domain.User{
					Role: domain.RoleModerator,
				},
			)

			c.Next()
		},
		RequireRole(
			domain.RoleModerator,
		),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleRejectsWrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.GET(
		"/test",
		func(c *gin.Context) {
			c.Set(
				CurrentUserContextKey,
				&domain.User{
					Role: domain.RoleGuest,
				},
			)

			c.Next()
		},
		RequireRole(
			domain.RoleModerator,
		),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected 403, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleRejectsMissingUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.GET(
		"/test",
		RequireRole(
			domain.RoleModerator,
		),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			recorder.Code,
		)
	}
}

func TestRequireRoleAllowsOneOfSeveralRoles(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.GET(
		"/test",
		func(c *gin.Context) {
			c.Set(
				CurrentUserContextKey,
				&domain.User{
					Role: domain.RoleHotel,
				},
			)

			c.Next()
		},
		RequireRole(
			domain.RoleHotel,
			domain.RoleModerator,
		),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d",
			recorder.Code,
		)
	}
}
