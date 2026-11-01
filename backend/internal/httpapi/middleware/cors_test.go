package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsConfigureOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cors := NewCORS(
		[]string{
			"http://localhost:5173",
		},
	)

	router := gin.New()
	router.Use(cors.Middleware())

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

	req.Header.Set(
		"Origin",
		"http://localhost:5173",
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

	if got := recorder.Header().Get(
		"Access-Control-Allow-Origin",
	); got != "http://localhost:5173" {
		t.Fatalf(
			"unexpected allow origin: %q",
			got,
		)
	}

	if got := recorder.Header().Get(
		"Access-Control-Allow-Credentials",
	); got != "true" {
		t.Fatalf(
			"expected credentials=true, got %q",
			got,
		)
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cors := NewCORS(
		[]string{
			"http://localhost:5173",
		},
	)

	router := gin.New()
	router.Use(cors.Middleware())

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

	req.Header.Set(
		"Origin",
		"https://evil.example",
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

func TestCORSAllowsRequestWithoutOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cors := NewCORS(
		[]string{
			"http://localhost:5173",
		},
	)

	router := gin.New()
	router.Use(cors.Middleware())

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

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d",
			recorder.Code,
		)
	}
}
