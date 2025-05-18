package validate

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ParseReqBody parses and validates incoming JSON request body into the given struct.
//
// - Uses a pre-initialized global validator instance to avoid unnecessary allocations.
// - First attempts to bind the JSON body to the provided struct via ctx.ShouldBindJSON.
// - Then validates the struct using go-playground/validator.
// - Returns detailed parsing or validation errors, if any.
//
// Note: validator.New() is created only once globally to leverage caching
// of struct metadata and improve performance.
var v = validator.New()

func ParseReqBody(ctx *gin.Context, reqBody interface{}) error {

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	if err := v.Struct(reqBody); err != nil {
		return fmt.Errorf("failed to validate request body: %w", err)
	}

	return nil
}
