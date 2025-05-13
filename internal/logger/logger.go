package logger

import (
	"context"
	"log/slog"
	"os"
)

func getArgs(args []slog.Attr) []any {
	var res []any
	for _, a := range args {
		res = append(res, a.Key, a.Value)
	}
	return res
}

type сtxValueKey struct{}

func getAttrs(ctx context.Context) []slog.Attr {
	av := ctx.Value(сtxValueKey{})
	if av == nil {
		return nil
	}
	return av.([]slog.Attr)
}

func mergeAttrs(ctx context.Context, attrs []slog.Attr) []slog.Attr {
	existing := getAttrs(ctx)
	return append(existing, attrs...)
}

func WithAttr(ctx context.Context, attrs ...slog.Attr) context.Context {
	merged := mergeAttrs(ctx, attrs)
	return context.WithValue(ctx, сtxValueKey{}, merged)
}

// Init initializes the logger before main
func init() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	l := slog.New(h)
	slog.SetDefault(l)
}

func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	args := getArgs(mergeAttrs(ctx, attrs))
	slog.Default().InfoContext(ctx, msg, args...)

}

func Error(ctx context.Context, err error, attrs ...slog.Attr) {
	args := getArgs(mergeAttrs(ctx, attrs))
	slog.Default().ErrorContext(ctx, err.Error(), args...)

}

func Panic(ctx context.Context, err error, attrs ...slog.Attr) {
	Error(ctx, err, attrs...)
	panic(err)
}

func Fatal(ctx context.Context, err error, attrs ...slog.Attr) {
	Error(ctx, err, attrs...)
	os.Exit(1)
}
