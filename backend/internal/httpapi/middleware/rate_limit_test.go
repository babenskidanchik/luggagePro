package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimitBlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewIPRateLimiter(
		2,
		time.Minute,
	)

	router := gin.New()

	router.GET(
		"/test",
		limiter.Middleware(),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		req.RemoteAddr = "127.0.0.1:12345"

		recorder := httptest.NewRecorder()

		router.ServeHTTP(
			recorder,
			req,
		)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf(
				"request %d: expected 204, got %d",
				i+1,
				recorder.Code,
			)
		}
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	req.RemoteAddr = "127.0.0.1:12345"

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected 429, got %d",
			recorder.Code,
		)
	}
}
