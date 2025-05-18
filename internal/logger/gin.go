package logger

import (
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
)

const ginLoggerKey = "loggerAttrs"

// ginGetLoggerAttr retrieves structured logging attributes from Gin context
func ginGetLoggerAttr(ctx *gin.Context) []slog.Attr {

	val, exists := ctx.Get(ginLoggerKey)
	if !exists {
		return nil
	}

	attrs, ok := val.([]slog.Attr)
	if !ok {
		return nil
	}
	return attrs

}

// ginMergeAttr merges existing Gin context attributes with new ones
func ginMergeAttr(ctx *gin.Context, attr []slog.Attr) []slog.Attr {
	existingAttr := ginGetLoggerAttr(ctx)
	return append(existingAttr, attr...)
}

// GinSetLoggerAttr sets structured logging attributes into the Gin context
func GinSetLoggerAttr(ctx *gin.Context, attrs ...slog.Attr) {
	attr := ginMergeAttr(ctx, attrs)
	ctx.Set(ginLoggerKey, attr)
}

// EnrichContextFromGin copies logging attributes from Gin context into standard context.Context
func EnrichContextFromGin(baseCtx context.Context, ginCtx *gin.Context) context.Context {
	val, ok := ginCtx.Get("loggerAttrs")
	if !ok {
		return baseCtx
	}
	attrs, ok := val.([]slog.Attr)
	if !ok {
		return baseCtx
	}
	return WithAttr(baseCtx, attrs...)
}
