package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func Logging(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Int("status", c.Writer.Status()).
			Dur("duration", time.Since(start)).
			Str("request_id", GetRequestID(c.Request.Context())).
			Str("remote_addr", c.Request.RemoteAddr).
			Str("user_agent", c.Request.UserAgent()).
			Msg("request completed")
	}
}
