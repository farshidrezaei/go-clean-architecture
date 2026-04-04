package common

import (
	"net/http"

	"clean_architecture/internal/usecase/shared"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "userID"
	ContextUserEmailKey = "userEmail"
	ContextUserRoleKey  = "userRole"
	ContextRequestIDKey = "requestID"
)

func UserID(c *gin.Context) string {
	return c.GetString(ContextUserIDKey)
}

func UserRole(c *gin.Context) string {
	return c.GetString(ContextUserRoleKey)
}

func Error(c *gin.Context, err error) {
	c.JSON(shared.HTTPStatus(err), gin.H{
		"error": gin.H{
			"code":       shared.Code(err),
			"message":    shared.SafeMessage(err),
			"request_id": c.GetString(ContextRequestIDKey),
		},
	})
}

func ValidationError(c *gin.Context, message string, details any) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": gin.H{
			"code":       "validation_error",
			"message":    message,
			"details":    details,
			"request_id": c.GetString(ContextRequestIDKey),
		},
	})
}

func Created(c *gin.Context, payload any) {
	c.JSON(http.StatusCreated, payload)
}

func OK(c *gin.Context, payload any) {
	c.JSON(http.StatusOK, payload)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
