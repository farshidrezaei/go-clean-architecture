package ginadapter

import (
	"log/slog"
	"time"

	"clean_architecture/internal/infrastructure/logger"
	"clean_architecture/internal/interface/http/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestContext(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(common.ContextRequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		requestLogger := logger.WithRequest(base, requestID)
		c.Request = c.Request.WithContext(logger.IntoContext(c.Request.Context(), requestLogger))
		c.Next()
	}
}

func RequestLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		ctx := newContext(c)
		reqLogger := logger.FromContext(c.Request.Context(), base)
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		if route := ctx.Route(); route != "" {
			attrs = append(attrs, "route", route)
		}
		if requestID := ctx.GetString(common.ContextRequestIDKey); requestID != "" {
			attrs = append(attrs, "request_id", requestID)
		}
		if userID := ctx.GetString(common.ContextUserIDKey); userID != "" {
			attrs = append(attrs, "user_id", userID)
		}
		if errs := ctx.Errors(); errs != "" {
			attrs = append(attrs, "errors", errs)
		}

		reqLogger.Info("http_request", attrs...)
	}
}
