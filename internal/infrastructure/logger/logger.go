package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

func New(env, level string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
	}

	if strings.EqualFold(env, "production") {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func WithRequest(logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With("request_id", requestID)
}

func FromContext(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if ctx == nil {
		return fallback
	}
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return fallback
}

func IntoContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

type contextKey struct{}

func parseLevel(level string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
