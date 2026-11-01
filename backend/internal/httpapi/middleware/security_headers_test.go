package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders(false))

	router.GET(
		"/api/v1/auth/current",
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/auth/current",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if got := recorder.Header().Get(
		"X-Content-Type-Options",
	); got != "nosniff" {
		t.Fatalf(
			"unexpected X-Content-Type-Options: %q",
			got,
		)
	}

	if got := recorder.Header().Get(
		"X-Frame-Options",
	); got != "DENY" {
		t.Fatalf(
			"unexpected X-Frame-Options: %q",
			got,
		)
	}

	if got := recorder.Header().Get(
		"Referrer-Policy",
	); got != "no-referrer" {
		t.Fatalf(
			"unexpected Referrer-Policy: %q",
			got,
		)
	}

	if got := recorder.Header().Get(
		"Cache-Control",
	); got != "no-store" {
		t.Fatalf(
			"expected auth Cache-Control=no-store, got %q",
			got,
		)
	}
}

func TestSecurityHeadersAddsHSTSInProduction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(
		SecurityHeaders(true),
	)

	router.GET(
		"/test",
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

	if got := recorder.Header().Get(
		"Strict-Transport-Security",
	); got == "" {
		t.Fatal(
			"expected HSTS header in production",
		)
	}
}
