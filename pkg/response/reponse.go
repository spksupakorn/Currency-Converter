package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	TraceID interface{} `json:"trace_id,omitempty"`
}

func WithStatus(c *gin.Context, status int, code, message string, details interface{}) {
	traceID, _ := c.Get("trace_id")
	c.JSON(status, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
		TraceID: traceID,
	})
}

func ValidationError(c *gin.Context, code string, err interface{}) {
	WithStatus(c, http.StatusBadRequest, code, "validation error", err)
}

func BadRequest(c *gin.Context, code string, message string) {
	WithStatus(c, http.StatusBadRequest, code, message, nil)
}

func Unauthorized(c *gin.Context, code string, message string) {
	WithStatus(c, http.StatusUnauthorized, code, message, nil)
}

func Forbidden(c *gin.Context, code string, message string) {
	WithStatus(c, http.StatusForbidden, code, message, nil)
}

func NotFound(c *gin.Context, code string, message string) {
	WithStatus(c, http.StatusNotFound, code, message, nil)
}

func TooManyRequests(c *gin.Context, code string, message string) {
	WithStatus(c, http.StatusTooManyRequests, code, message, nil)
}

func InternalError(c *gin.Context, code string, message string) {
	WithStatus(c, http.StatusInternalServerError, code, message, nil)
}
