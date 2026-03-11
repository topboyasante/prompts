package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Details   map[string]string `json:"details,omitempty"`
}

func RespondError(c *gin.Context, status int, code, message string) {
	if status < 100 {
		status = http.StatusInternalServerError
	}
	c.JSON(status, ErrorResponse{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			RequestID: RequestIDFromContext(c),
		},
	})
}

func RespondValidationError(c *gin.Context, details map[string]string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorDetail{
			Code:      "VALIDATION_FAILED",
			Message:   "validation failed",
			RequestID: RequestIDFromContext(c),
			Details:   details,
		},
	})
}

func RespondJSON(c *gin.Context, status int, v any) {
	c.JSON(status, v)
}
