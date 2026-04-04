package middleware

import (
	"context"
	"net/http"
	"time"

	"clean_architecture/internal/infrastructure/ratelimit"
	"clean_architecture/internal/interface/http/common"
	"github.com/gin-gonic/gin"
)

func LoginRateLimit(service ratelimit.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + ":login"
		allowed, err := service.Allow(context.Background(), key, 5, time.Minute)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal_error", "message": "rate limiter failed", "request_id": c.GetString(common.ContextRequestIDKey)}})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "rate_limited", "message": "too many login attempts", "request_id": c.GetString(common.ContextRequestIDKey)}})
			return
		}
		c.Next()
	}
}

func RefreshRateLimit(service ratelimit.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + ":refresh"
		allowed, err := service.Allow(context.Background(), key, 10, time.Minute)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal_error", "message": "rate limiter failed", "request_id": c.GetString(common.ContextRequestIDKey)}})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "rate_limited", "message": "too many refresh attempts", "request_id": c.GetString(common.ContextRequestIDKey)}})
			return
		}
		c.Next()
	}
}
