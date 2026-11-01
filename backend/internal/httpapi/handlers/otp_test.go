package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type fakeHTTPOTPService struct {
	requestErr error
	verifyErr  error
	token      string

	requestPhone string
	verifyPhone  string
	verifyCode   string
}

func (f *fakeHTTPOTPService) RequestOTP(
	ctx context.Context,
	phone string,
) error {
	f.requestPhone = phone

	return f.requestErr
}

func (f *fakeHTTPOTPService) VerifyOTP(
	ctx context.Context,
	phone string,
	code string,
	userAgent *string,
	ipAddress *string,
) (string, error) {
	f.verifyPhone = phone
	f.verifyCode = code

	if f.verifyErr != nil {
		return "", f.verifyErr
	}

	return f.token, nil
}

func TestOTPRequestSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeHTTPOTPService{}

	handler := NewOTPHandler(
		service,
		true,
	)

	router := gin.New()

	router.POST(
		"/otp/request",
		handler.Request,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/otp/request",
		strings.NewReader(
			`{"phone":"+995555123456"}`,
		),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if service.requestPhone != "+995555123456" {
		t.Fatalf(
			"unexpected phone: %q",
			service.requestPhone,
		)
	}
}

func TestOTPRequestRejectsUnknownFields(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHTTPOTPService{}

	handler := NewOTPHandler(
		service,
		true,
	)

	router := gin.New()

	router.POST(
		"/otp/request",
		handler.Request,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/otp/request",
		strings.NewReader(
			`{
				"phone":"+995555123456",
				"role":"moderator"
			}`,
		),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400, got %d",
			response.Code,
		)
	}
}

func TestOTPRequestCooldown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeHTTPOTPService{
		requestErr: domain.ErrOTPCooldown,
	}

	handler := NewOTPHandler(
		service,
		true,
	)

	router := gin.New()

	router.POST(
		"/otp/request",
		handler.Request,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/otp/request",
		strings.NewReader(
			`{"phone":"+995555123456"}`,
		),
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected 429, got %d",
			response.Code,
		)
	}
}

func TestOTPVerifySetsSecureCookie(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHTTPOTPService{
		token: "test-session-token",
	}

	handler := NewOTPHandler(
		service,
		true,
	)

	router := gin.New()

	router.POST(
		"/otp/verify",
		handler.Verify,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/otp/verify",
		strings.NewReader(
			`{
				"phone":"+995555123456",
				"code":"583291"
			}`,
		),
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	cookies := response.Result().Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected one cookie, got %d",
			len(cookies),
		)
	}

	cookie := cookies[0]

	if cookie.Name != "session_token" {
		t.Fatalf(
			"unexpected cookie name: %q",
			cookie.Name,
		)
	}

	if !cookie.HttpOnly {
		t.Fatal(
			"expected HttpOnly cookie",
		)
	}

	if !cookie.Secure {
		t.Fatal(
			"expected Secure cookie",
		)
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf(
			"expected SameSite=Lax, got %v",
			cookie.SameSite,
		)
	}

	if response.Body.Len() != 0 {
		t.Fatal(
			"session token must not appear in response body",
		)
	}
}

func TestOTPVerifyRejectsInvalidCode(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHTTPOTPService{
		verifyErr: domain.ErrOTPInvalid,
	}

	handler := NewOTPHandler(
		service,
		true,
	)

	router := gin.New()

	router.POST(
		"/otp/verify",
		handler.Verify,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/otp/verify",
		strings.NewReader(
			`{
				"phone":"+995555123456",
				"code":"000000"
			}`,
		),
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			response.Code,
		)
	}

	if strings.Contains(
		response.Body.String(),
		"000000",
	) {
		t.Fatal(
			"OTP must not be echoed in response",
		)
	}
}

func TestOTPVerifyInternalErrorDoesNotLeak(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	service := &fakeHTTPOTPService{
		verifyErr: errors.New(
			"postgres password=secret internal error",
		),
	}

	handler := NewOTPHandler(
		service,
		true,
	)

	router := gin.New()

	router.POST(
		"/otp/verify",
		handler.Verify,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/otp/verify",
		strings.NewReader(
			`{
				"phone":"+995555123456",
				"code":"123456"
			}`,
		),
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d",
			response.Code,
		)
	}

	if strings.Contains(
		response.Body.String(),
		"postgres",
	) {
		t.Fatal(
			"internal error leaked to client",
		)
	}
}
