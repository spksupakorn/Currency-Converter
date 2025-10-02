package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
)

func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		traceID, _ := c.Get("trace_id")
		c.Next()
		lat := time.Since(start)
		log.Info("http_request",
			logger.Fields{
				"method":      c.Request.Method,
				"path":        c.FullPath(),
				"status":      c.Writer.Status(),
				"duration_ms": lat.Milliseconds(),
				"ip":          c.ClientIP(),
				"ua":          c.Request.UserAgent(),
				"trace_id":    traceID,
			},
		)
	}
}
