package response

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// AbortWithError is a utility function to abort the request
// and add error to context to display in logging middleware.
func AbortWithError(ctx *gin.Context, statusCode int, err error) {
	ctx.Error(err)
	ctx.AbortWithStatus(statusCode)
}

// AbortWithErrorJSON is a utility function to abort the request
// and add error to context to display in logging middleware.
// Return secure message to user.
func AbortWithErrorJSON(ctx *gin.Context, statusCode int, err error, usrErrMsg string) {
	err = fmt.Errorf("%d: %w", statusCode, err)
	ctx.Error(err)

	ctx.AbortWithStatusJSON(statusCode, ErrorResponse{
		Error: usrErrMsg,
		Code:  fmt.Sprintf("ERR_%d", statusCode),
	})
}
