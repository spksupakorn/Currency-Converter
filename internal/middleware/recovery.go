package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
	"github.com/spksupakorn/Currency-Converter/pkg/response"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				traceID, _ := c.Get("trace_id")
				log.Error("panic recovered", logger.Fields{
					"panic":    rec,
					"trace_id": traceID,
				})
				response.WithStatus(c, http.StatusInternalServerError, "internal_error", "internal server error", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
