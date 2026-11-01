package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCSRFTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	csrf := NewCSRF(false)

	router := gin.New()
	router.Use(csrf.Middleware())

	router.GET(
		"/test",
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	router.POST(
		"/test",
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	return router
}

func TestCSRFIssuesTokenOnSafeRequest(t *testing.T) {
	router := newCSRFTestRouter()

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

	var found bool

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == CSRFCookieName &&
			cookie.Value != "" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected CSRF cookie")
	}
}

func TestCSRFRejectsMissingTokenForCookieAuth(t *testing.T) {
	router := newCSRFTestRouter()

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		nil,
	)

	req.AddCookie(
		&http.Cookie{
			Name:  SessionCookieName,
			Value: "test-session",
		},
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

func TestCSRFAcceptsMatchingToken(t *testing.T) {
	router := newCSRFTestRouter()

	const token = "token-csrf-token"

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		nil,
	)

	req.AddCookie(
		&http.Cookie{
			Name:  SessionCookieName,
			Value: "test-session",
		},
	)

	req.AddCookie(
		&http.Cookie{
			Name:  CSRFCookieName,
			Value: token,
		},
	)

	req.Header.Set(
		CSRFHeaderName,
		token,
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

func TestCSRFRejectsWrongToken(t *testing.T) {
	router := newCSRFTestRouter()

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		nil,
	)

	req.AddCookie(
		&http.Cookie{
			Name:  SessionCookieName,
			Value: "test-session",
		},
	)

	req.AddCookie(
		&http.Cookie{
			Name:  CSRFCookieName,
			Value: "correct-token",
		},
	)

	req.Header.Set(
		CSRFHeaderName,
		"wrong-token",
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

func TestCSRFAllowsBearerAuth(t *testing.T) {
	router := newCSRFTestRouter()

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer moblie-session-token",
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
