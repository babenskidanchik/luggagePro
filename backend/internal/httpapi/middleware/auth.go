package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName     = "session_token"
	CurrentUserContextKey = "current_user"
)

type AuthService interface {
	CurrentUser(
		ctx context.Context,
		token string,
	) (*domain.User, error)
}

func RequireAuth(service AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := ExtractSessionToken(c)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "authentication required",
				},
			)
			return
		}

		user, err := service.CurrentUser(
			c.Request.Context(),
			token,
		)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid or expired session",
				},
			)
			return
		}

		c.Set(CurrentUserContextKey, user)

		c.Next()
	}
}

func ExtractSessionToken(c *gin.Context) (string, bool) {
	authorization := strings.TrimSpace(
		c.GetHeader("Authorization"),
	)

	if authorization != "" {
		parts := strings.Fields(authorization)

		if len(parts) != 2 {
			return "", false
		}

		if !strings.EqualFold(parts[0], "Bearer") {
			return "", false
		}

		token := strings.TrimSpace(parts[1])

		if token == "" {
			return "", false
		}

		return token, true
	}

	token, err := c.Cookie(SessionCookieName)
	if err != nil {
		return "", false
	}

	token = strings.TrimSpace(token)

	if token == "" {
		return "", false
	}

	return token, true
}

func CurrentUser(c *gin.Context) (*domain.User, bool) {
	value, exists := c.Get(CurrentUserContextKey)
	if !exists {
		return nil, false
	}

	user, ok := value.(*domain.User)

	return user, ok
}
