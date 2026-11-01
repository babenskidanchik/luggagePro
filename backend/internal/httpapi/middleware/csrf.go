package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CSRFCookieName = "csrf-token"
	CSRFHeaderName = "X-CSRF-Token"
)

type CSRF struct {
	secureCookie bool
}

func NewCSRF(
	secureCookie bool,
) *CSRF {
	return &CSRF{
		secureCookie: secureCookie,
	}
}

func (m *CSRF) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) {
			if err := m.ensureToken(c); err != nil {
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"error": "internal server error",
					},
				)
				return
			}

			c.Next()
			return
		}

		if usesBearerAuth(c) {
			c.Next()
			return
		}

		if isOTPBootstrapRoute(c) {
			c.Next()
			return
		}

		if _, err := c.Cookie(SessionCookieName); err != nil {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"error": "csrf validation failed",
				},
			)
			return
		}

		headerToken := strings.TrimSpace(
			c.GetHeader(CSRFHeaderName),
		)

		if headerToken == "" ||
			!hmac.Equal(
				[]byte(cookieToken),
				[]byte(headerToken),
			) {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"error": "csrf validation failed",
				},
			)
			return
		}

		c.Next()
	}
}

func (m *CSRF) ensureToken(
	c *gin.Context,
) error {
	if token, err := c.Cookie(CSRFCookieName); err == nil &&
		token != "" {
		return nil
	}

	token, err := generateCSRFToken()
	if err != nil {
		return err
	}

	http.SetCookie(
		c.Writer,
		&http.Cookie{
			Name:     CSRFCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: false,
			Secure:   m.secureCookie,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400,
		},
	)

	return nil
}

func generateCSRFToken() (
	string,
	error,
) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func isSafeMethod(
	method string,
) bool {
	switch method {
	case http.MethodGet,
		http.MethodHead,
		http.MethodOptions:
		return true
	default:
		return false
	}
}

func usesBearerAuth(
	c *gin.Context,
) bool {
	authHeader := strings.TrimSpace(
		c.GetHeader("Authorization"),
	)

	return strings.HasPrefix(
		authHeader,
		"Bearer ",
	)
}

func isOTPBootstrapRoute(
	c *gin.Context,
) bool {
	switch c.Request.URL.Path {
	case "/api/v1/auth/otp/request",
		"/api/v1/auth/otp/verify":
		return true
	default:
		return false
	}
}
