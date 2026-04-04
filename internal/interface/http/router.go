package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"clean_architecture/internal/infrastructure/health"
	appLogger "clean_architecture/internal/infrastructure/logger"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/handlers"
	"clean_architecture/internal/interface/http/middleware"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	UserHandler      *handlers.UserHandler
	PostHandler      *handlers.PostHandler
	CommentHandler   *handlers.CommentHandler
	AdminHandler     *handlers.AdminHandler
	AuthMiddleware   gin.HandlerFunc
	LoginRateLimit   gin.HandlerFunc
	RefreshRateLimit gin.HandlerFunc
	Logger           *slog.Logger
	HealthChecker    health.Checker
	Metrics          gin.HandlerFunc
	MetricsHandler   gin.HandlerFunc
	Tracing          gin.HandlerFunc
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	router := gin.New()
	if deps.LoginRateLimit == nil {
		deps.LoginRateLimit = func(c *gin.Context) { c.Next() }
	}
	if deps.RefreshRateLimit == nil {
		deps.RefreshRateLimit = func(c *gin.Context) { c.Next() }
	}
	router.Use(gin.Recovery())
	router.Use(middleware.RequestContext(deps.Logger))
	if deps.Tracing != nil {
		router.Use(deps.Tracing)
	}
	router.Use(requestLogger(deps.Logger))
	if deps.Metrics != nil {
		router.Use(deps.Metrics)
	}

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if deps.HealthChecker != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := deps.HealthChecker.Check(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	if deps.MetricsHandler != nil {
		router.GET("/metrics", deps.MetricsHandler)
	}
	router.GET("/openapi.yaml", openAPIHandler())
	router.GET("/docs", docsHandler())

	api := router.Group("/api/v1")
	api.POST("/auth/register", deps.UserHandler.Register)
	api.POST("/auth/login", deps.LoginRateLimit, deps.UserHandler.Login)
	api.POST("/auth/refresh", deps.RefreshRateLimit, deps.UserHandler.Refresh)
	api.POST("/auth/logout", deps.UserHandler.Logout)

	api.GET("/posts", deps.PostHandler.List)
	api.GET("/posts/:postID", deps.PostHandler.Get)
	api.GET("/posts/:postID/comments", deps.CommentHandler.List)

	protected := api.Group("/")
	protected.Use(deps.AuthMiddleware)
	protected.POST("/posts", deps.PostHandler.Create)
	protected.PUT("/posts/:postID", deps.PostHandler.Update)
	protected.DELETE("/posts/:postID", deps.PostHandler.Delete)
	protected.POST("/posts/:postID/likes", deps.PostHandler.Like)
	protected.POST("/posts/:postID/comments", deps.CommentHandler.Add)
	protected.PUT("/comments/:commentID", deps.CommentHandler.Update)
	protected.DELETE("/comments/:commentID", deps.CommentHandler.Delete)
	protected.POST("/auth/logout-all", deps.UserHandler.LogoutAll)
	protected.GET("/auth/sessions", deps.UserHandler.ListSessions)
	protected.POST("/auth/sessions/revoke", deps.UserHandler.RevokeSession)

	if deps.AdminHandler != nil {
		adminGroup := protected.Group("/admin")
		adminGroup.Use(middleware.RequireAdmin())
		adminGroup.GET("/users", deps.AdminHandler.ListUsers)
		adminGroup.PUT("/users/:userID/role", deps.AdminHandler.UpdateUserRole)
	}

	return router
}

func NewAuthMiddleware(tokens middleware.Authenticator) gin.HandlerFunc {
	return middleware.RequireAuth(tokens)
}

func requestLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger := appLogger.FromContext(c.Request.Context(), base)
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		if route := c.FullPath(); route != "" {
			attrs = append(attrs, "route", route)
		}
		if requestID := c.GetString(common.ContextRequestIDKey); requestID != "" {
			attrs = append(attrs, "request_id", requestID)
		}
		if userID := c.GetString(common.ContextUserIDKey); userID != "" {
			attrs = append(attrs, "user_id", userID)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		logger.Info("http_request", attrs...)
	}
}
