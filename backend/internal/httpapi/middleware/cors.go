package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CORS struct {
	allowedOrigins map[string]struct{}
}

func NewCORS(
	origins []string,
) *CORS {
	allowed := make(map[string]struct{}, len(origins))

	for _, origin := range origins {
		if origin == "" {
			continue
		}

		allowed[origin] = struct{}{}

	}

	return &CORS{
		allowedOrigins: allowed,
	}
}

func (m *CORS) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin == "" {
			c.Next()
			return
		}

		if _, ok := m.allowedOrigins[origin]; !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"error": "origin not allowed",
				},
			)
			return
		}

		c.Header(
			"Access-Control-Allow-Origin",
			origin,
		)

		c.Header(
			"Access-Control-Allow-Credentials",
			"true",
		)

		c.Header(
			"Vary",
			"Origin",
		)

		c.Header(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)

		c.Header(
			"Access-Control-Allow-Headers",
			"Authorization, Content-Type, X-CSRF-Token",
		)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(
				http.StatusNoContent,
			)
			return
		}

		c.Next()
	}
}
