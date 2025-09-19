package response

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteErrorJSON writes standardized JSON error and attaches err to gin.Context for logging middleware.
func WriteErrorJSON(ctx *gin.Context, statusCode int, err error, userMsg string) {
	if err != nil {
		ctx.Error(err)
	}
	ctx.AbortWithStatusJSON(statusCode, ErrorResponse{
		Error: userMsg,
	})
}
