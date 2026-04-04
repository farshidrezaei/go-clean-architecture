package middleware

import (
	"net/http"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/interface/http/common"
	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if common.UserRole(c) != string(entities.RoleAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "admin role is required", "request_id": c.GetString(common.ContextRequestIDKey)}})
			return
		}
		c.Next()
	}
}
