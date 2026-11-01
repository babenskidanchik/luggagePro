package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimitRejectsOversizedRequest(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.Use(
		BodyLimit(10),
	)

	router.POST(
		"/test",
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		strings.NewReader(
			"this body is definitely too large",
		),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(
			"expected 413, got %d",
			recorder.Code,
		)
	}
}

func TestBodyLimitAllowsSmallRequest(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.Use(
		BodyLimit(1024),
	)

	router.POST(
		"/test",
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		strings.NewReader(`{"ok":true}`),
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
