package validate

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// доволі масивна штука - там багато оптимізацій, кешування і т.д.
// правильно не створювати на кожен чих новий валідатор

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
