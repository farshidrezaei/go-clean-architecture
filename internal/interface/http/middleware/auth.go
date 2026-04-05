package middleware

import (
	"net/http"
	"strings"

	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
)

type Authenticator interface {
	ValidateToken(tokenString string) (subject string, email string, role string, err error)
}

func RequireAuth(tokens Authenticator) port.MiddlewareFunc {
	return func(ctx port.Context) {
		header := ctx.Header("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{"error": map[string]any{"code": "unauthorized", "message": "missing bearer token", "request_id": ctx.GetString(common.ContextRequestIDKey)}})
			return
		}

		rawToken := strings.TrimPrefix(header, "Bearer ")
		subject, email, role, err := tokens.ValidateToken(rawToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, map[string]any{"error": map[string]any{"code": "unauthorized", "message": "invalid token", "request_id": ctx.GetString(common.ContextRequestIDKey)}})
			return
		}

		ctx.Set(common.ContextUserIDKey, subject)
		ctx.Set(common.ContextUserEmailKey, email)
		ctx.Set(common.ContextUserRoleKey, role)
		ctx.Next()
	}
}
