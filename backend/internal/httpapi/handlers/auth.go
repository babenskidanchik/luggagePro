package handlers

import (
	"context"
	"net/http"

	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

func CurrentUser(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"id":    user.ID,
			"phone": user.Phone,
			"email": user.Email,
			"role":  user.Role,
		},
	)
}

type LogoutService interface {
	Logout(
		ctx context.Context,
		token string,
	) error
}

func Logout(service LogoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := middleware.ExtractSessionToken(c)
		if !ok {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "authentication required",
				},
			)
			return
		}

		if err := service.Logout(
			c.Request.Context(),
			token,
		); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "internal server error",
				},
			)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
