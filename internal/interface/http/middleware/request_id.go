package middleware

import (
	"log/slog"

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
