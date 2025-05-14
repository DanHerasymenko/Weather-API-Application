package logger

import (
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
)

const ginLoggerKey = "loggerAttrs"

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

func ginMergeAttr(ctx *gin.Context, attr []slog.Attr) []slog.Attr {
	existingAttr := ginGetLoggerAttr(ctx)
	return append(existingAttr, attr...)
}

func GinSetLoggerAttr(ctx *gin.Context, attrs ...slog.Attr) {
	attr := ginMergeAttr(ctx, attrs)
	ctx.Set(ginLoggerKey, attr)
}

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
