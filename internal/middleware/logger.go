package middleware

import (
	"Weather-API-Application/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"log/slog"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqID := ulid.Make().String()
		reqPath := ctx.Request.URL.Path
		reqMethod := ctx.Request.Method

		logger.GinSetLoggerAttr(
			ctx,
			slog.String("request_id", reqID),
			slog.String("request_path", reqPath),
			slog.String("method", reqMethod),
		)
		logger.Info(ctx.Request.Context(), "request start")

		startedAt := time.Now()
		ctx.Next()
		duration := time.Since(startedAt)

		logger.GinSetLoggerAttr(ctx, slog.Int64("duration_ms", duration.Milliseconds()))

		// If there were any errors during request handling, attach the last one
		if len(ctx.Errors) > 0 {
			for _, err := range ctx.Errors {
				logger.GinSetLoggerAttr(ctx, slog.String("resp_message", err.Error()))
			}
		}

		logger.GinSetLoggerAttr(ctx, slog.Int("status", ctx.Writer.Status()))

		ctxMerged := logger.EnrichContextFromGin(ctx.Request.Context(), ctx)

		logger.Info(ctxMerged, "request end")
	}
}
