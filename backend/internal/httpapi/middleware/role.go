package middleware

import (
	"net/http"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func RequireRole(
	allowedRoles ...domain.UserRole,
) gin.HandlerFunc {
	allowed := make(
		map[domain.UserRole]struct{},
		len(allowedRoles),
	)

	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		value, exists := c.Get(
			CurrentUserContextKey,
		)

		if !exists {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "authentication required",
				},
			)
			return
		}

		user, ok := value.(*domain.User)
		if !ok || user == nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "authentication required",
				},
			)
			return
		}

		if _, ok := allowed[user.Role]; !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"error": "forbidden",
				},
			)
			return
		}

		c.Next()
	}
}
