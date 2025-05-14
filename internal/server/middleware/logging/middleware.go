package logging

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"log/slog"
	"time"
)

type Middleware struct {
	cfg *config.Config
}

func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{
		cfg: cfg,
	}
}

func (m *Middleware) Handle(ctx *gin.Context) {

	reqID := ulid.Make().String()
	reqPath := ctx.Request.URL.Path

	logger.GinSetLoggerAttr(ctx, slog.String("request_id", reqID), slog.String("request_path", reqPath))
	logger.Info(ctx.Request.Context(), "request start")

	startedAt := time.Now()
	ctx.Next()
	duration := time.Since(startedAt)

	logger.GinSetLoggerAttr(ctx, slog.Int64("duration_ms", duration.Milliseconds()))

	if len(ctx.Errors) > 0 {
		for _, err := range ctx.Errors {
			logger.GinSetLoggerAttr(ctx, slog.String("resp_message", err.Error()))
		}
	}

	logger.GinSetLoggerAttr(ctx, slog.Int("status", ctx.Writer.Status()))

	ctxMerged := logger.EnrichContextFromGin(ctx.Request.Context(), ctx)

	logger.Info(ctxMerged, "request end")

}
