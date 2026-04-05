package middleware

import (
	"net/http"
	"time"

	"clean_architecture/internal/infrastructure/ratelimit"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
)

func LoginRateLimit(service ratelimit.Service) port.MiddlewareFunc {
	return func(ctx port.Context) {
		key := ctx.ClientIP() + ":login"
		allowed, err := service.Allow(ctx.Context(), key, 5, time.Minute)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{"error": map[string]any{"code": "internal_error", "message": "rate limiter failed", "request_id": ctx.GetString(common.ContextRequestIDKey)}})
			return
		}
		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]any{"error": map[string]any{"code": "rate_limited", "message": "too many login attempts", "request_id": ctx.GetString(common.ContextRequestIDKey)}})
			return
		}
		ctx.Next()
	}
}

func RefreshRateLimit(service ratelimit.Service) port.MiddlewareFunc {
	return func(ctx port.Context) {
		key := ctx.ClientIP() + ":refresh"
		allowed, err := service.Allow(ctx.Context(), key, 10, time.Minute)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, map[string]any{"error": map[string]any{"code": "internal_error", "message": "rate limiter failed", "request_id": ctx.GetString(common.ContextRequestIDKey)}})
			return
		}
		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]any{"error": map[string]any{"code": "rate_limited", "message": "too many refresh attempts", "request_id": ctx.GetString(common.ContextRequestIDKey)}})
			return
		}
		ctx.Next()
	}
}
