package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

const maxOTPRequestBodyBytes = 1024

type OTPService interface {
	RequestOTP(
		ctx context.Context,
		phone string,
	) error

	VerifyOTP(
		ctx context.Context,
		phone string,
		code string,
		userAgent *string,
		ipAddress *string,
	) (string, error)
}

type OTPHandler struct {
	service      OTPService
	secureCookie bool
}

func NewOTPHandler(
	service OTPService,
	secureCookie bool,
) *OTPHandler {
	return &OTPHandler{
		service:      service,
		secureCookie: secureCookie,
	}
}

type requestOTPBody struct {
	Phone string `json:"phone"`
}

type verifyOTPBody struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func (h *OTPHandler) Request(c *gin.Context) {
	var body requestOTPBody

	if err := decodeStrictJSON(c, &body); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request"},
		)
		return
	}

	err := h.service.RequestOTP(
		c.Request.Context(),
		body.Phone,
	)

	switch {
	case err == nil:
		c.Status(http.StatusNoContent)

	case errors.Is(err, domain.ErrInvalidPhone):
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid phone"},
		)

	case errors.Is(err, domain.ErrOTPCooldown):
		c.JSON(
			http.StatusTooManyRequests,
			gin.H{
				"error": "please wait before requesting another code",
			},
		)

	default:
		log.Printf("request OTP failed: %v", err)

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
	}
}

func (h *OTPHandler) Verify(c *gin.Context) {
	var body verifyOTPBody

	if err := decodeStrictJSON(c, &body); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request"},
		)
		return
	}

	userAgent := optionalString(
		c.Request.UserAgent(),
	)

	ipAddress := optionalString(
		c.ClientIP(),
	)

	token, err := h.service.VerifyOTP(
		c.Request.Context(),
		body.Phone,
		body.Code,
		userAgent,
		ipAddress,
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPhone):
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid phone"},
			)

		case errors.Is(err, domain.ErrOTPInvalid),
			errors.Is(err, domain.ErrOTPExpired),
			errors.Is(err, domain.ErrOTPNotFound),
			errors.Is(err, domain.ErrOTPTooManyAttempts):

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid or expired code",
				},
			)

		case errors.Is(err, domain.ErrPhoneInUse),
			errors.Is(err, domain.ErrUserInactive):

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "unable to authenticate",
				},
			)

		default:
			log.Printf("verify OTP failed: %v", err)

			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "internal server error",
				},
			)
		}

		return
	}

	http.SetCookie(
		c.Writer,
		&http.Cookie{
			Name:     middleware.SessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.secureCookie,
			SameSite: http.SameSiteLaxMode,
		},
	)

	c.Status(http.StatusNoContent)
}

func decodeStrictJSON(
	c *gin.Context,
	destination any,
) error {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxOTPRequestBodyBytes,
	)

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New(
			"request must contain exactly one JSON object",
		)
	}

	return nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	return &value
}
