package middleware

import (
	"net/http"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
)

func RequireAdmin() port.MiddlewareFunc {
	return func(ctx port.Context) {
		if common.UserRole(ctx) != string(entities.RoleAdmin) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, map[string]any{"error": map[string]any{"code": "forbidden", "message": "admin role is required", "request_id": ctx.GetString(common.ContextRequestIDKey)}})
			return
		}
		ctx.Next()
	}
}
