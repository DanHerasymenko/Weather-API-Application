package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

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

	ctx.AbortWithStatusJSON(statusCode, gin.H{
		"error": usrErrMsg,
	})
}
