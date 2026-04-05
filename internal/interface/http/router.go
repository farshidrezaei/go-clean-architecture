package http

import (
	"context"
	"net/http"
	"time"

	ginadapter "clean_architecture/internal/interface/http/adapter/gin"
	"clean_architecture/internal/interface/http/handlers"
	"clean_architecture/internal/interface/http/middleware"
	"clean_architecture/internal/interface/http/port"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	UserHandler      *handlers.UserHandler
	PostHandler      *handlers.PostHandler
	CommentHandler   *handlers.CommentHandler
	AdminHandler     *handlers.AdminHandler
	AuthMiddleware   port.MiddlewareFunc
	LoginRateLimit   port.MiddlewareFunc
	RefreshRateLimit port.MiddlewareFunc
	RequestContext   gin.HandlerFunc
	RequestLogger    gin.HandlerFunc
	HealthChecker    port.HealthChecker
	Metrics          port.MiddlewareFunc
	MetricsHandler   gin.HandlerFunc
	Tracing          port.MiddlewareFunc
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	router := gin.New()
	if deps.LoginRateLimit == nil {
		deps.LoginRateLimit = func(ctx port.Context) { ctx.Next() }
	}
	if deps.RefreshRateLimit == nil {
		deps.RefreshRateLimit = func(ctx port.Context) { ctx.Next() }
	}
	router.Use(gin.Recovery())
	if deps.RequestContext != nil {
		router.Use(deps.RequestContext)
	}
	if deps.Tracing != nil {
		router.Use(ginadapter.Middleware(deps.Tracing))
	}
	if deps.RequestLogger != nil {
		router.Use(deps.RequestLogger)
	}
	if deps.Metrics != nil {
		router.Use(ginadapter.Middleware(deps.Metrics))
	}

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"status": "ok"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if deps.HealthChecker != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := deps.HealthChecker.Check(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, map[string]any{"status": "degraded", "error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, map[string]any{"status": "ready"})
	})
	if deps.MetricsHandler != nil {
		router.GET("/metrics", deps.MetricsHandler)
	}
	router.GET("/openapi.yaml", openAPIHandler())
	router.GET("/docs", docsHandler())

	api := router.Group("/api/v1")
	api.POST("/auth/register", ginadapter.Handler(deps.UserHandler.Register))
	api.POST("/auth/login", ginadapter.Middleware(deps.LoginRateLimit), ginadapter.Handler(deps.UserHandler.Login))
	api.POST("/auth/refresh", ginadapter.Middleware(deps.RefreshRateLimit), ginadapter.Handler(deps.UserHandler.Refresh))
	api.POST("/auth/logout", ginadapter.Handler(deps.UserHandler.Logout))

	api.GET("/posts", ginadapter.Handler(deps.PostHandler.List))
	api.GET("/posts/:postID", ginadapter.Handler(deps.PostHandler.Get))
	api.GET("/posts/:postID/comments", ginadapter.Handler(deps.CommentHandler.List))

	protected := api.Group("/")
	if deps.AuthMiddleware != nil {
		protected.Use(ginadapter.Middleware(deps.AuthMiddleware))
	}
	protected.POST("/posts", ginadapter.Handler(deps.PostHandler.Create))
	protected.PUT("/posts/:postID", ginadapter.Handler(deps.PostHandler.Update))
	protected.DELETE("/posts/:postID", ginadapter.Handler(deps.PostHandler.Delete))
	protected.POST("/posts/:postID/likes", ginadapter.Handler(deps.PostHandler.Like))
	protected.POST("/posts/:postID/comments", ginadapter.Handler(deps.CommentHandler.Add))
	protected.PUT("/comments/:commentID", ginadapter.Handler(deps.CommentHandler.Update))
	protected.DELETE("/comments/:commentID", ginadapter.Handler(deps.CommentHandler.Delete))
	protected.POST("/auth/logout-all", ginadapter.Handler(deps.UserHandler.LogoutAll))
	protected.GET("/auth/sessions", ginadapter.Handler(deps.UserHandler.ListSessions))
	protected.POST("/auth/sessions/revoke", ginadapter.Handler(deps.UserHandler.RevokeSession))

	if deps.AdminHandler != nil {
		adminGroup := protected.Group("/admin")
		adminGroup.Use(ginadapter.Middleware(middleware.RequireAdmin()))
		adminGroup.GET("/users", ginadapter.Handler(deps.AdminHandler.ListUsers))
		adminGroup.PUT("/users/:userID/role", ginadapter.Handler(deps.AdminHandler.UpdateUserRole))
	}

	return router
}

func NewAuthMiddleware(tokens middleware.Authenticator) port.MiddlewareFunc {
	return middleware.RequireAuth(tokens)
}
