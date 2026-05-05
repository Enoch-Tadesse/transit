package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/henok/transit-backend/internal/delivery/http/response"
	"github.com/rs/zerolog"
)

func Recover(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error().
					Interface("panic", rec).
					Bytes("stack", debug.Stack()).
					Str("path", c.Request.URL.Path).
					Msg("handler panic recovered")
				response.WriteError(c, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
				c.Abort()
			}
		}()
		c.Next()
	}
}
