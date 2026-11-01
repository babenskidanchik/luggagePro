package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders(
	production bool,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header(
			"X-Content-Type-Options",
			"nosniff",
		)

		c.Header(
			"X-Frame-Options",
			"DENY",
		)

		c.Header(
			"Referrer-Policy",
			"no-referrer",
		)

		c.Header(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=()",
		)

		if production {
			c.Header(
				"Strict-Transport-Security",
				"max-age=315360000; includeSubDomains",
			)
		}

		if strings.HasPrefix(
			c.Request.URL.Path,
			"/api/v1/auth/",
		) {
			c.Header(
				"Cache-Control",
				"no-store",
			)

			c.Header(
				"Pragma",
				"no-cache",
			)
		}

		c.Next()
	}
}
