package middleware

import (
	"net/http"
	"strings"

	"clean_architecture/internal/interface/http/common"
	"github.com/gin-gonic/gin"
)

type Authenticator interface {
	ValidateToken(tokenString string) (subject string, email string, role string, err error)
}

func RequireAuth(tokens Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "missing bearer token", "request_id": c.GetString(common.ContextRequestIDKey)}})
			return
		}

		rawToken := strings.TrimPrefix(header, "Bearer ")
		subject, email, role, err := tokens.ValidateToken(rawToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "invalid token", "request_id": c.GetString(common.ContextRequestIDKey)}})
			return
		}

		c.Set(common.ContextUserIDKey, subject)
		c.Set(common.ContextUserEmailKey, email)
		c.Set(common.ContextUserRoleKey, role)
		c.Next()
	}
}
