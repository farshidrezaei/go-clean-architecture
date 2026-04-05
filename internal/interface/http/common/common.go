package common

import (
	"net/http"

	"clean_architecture/internal/interface/http/port"
	"clean_architecture/internal/usecase/shared"
)

const (
	ContextUserIDKey    = "userID"
	ContextUserEmailKey = "userEmail"
	ContextUserRoleKey  = "userRole"
	ContextRequestIDKey = "requestID"
)

func UserID(ctx port.Context) string {
	return ctx.GetString(ContextUserIDKey)
}

func UserRole(ctx port.Context) string {
	return ctx.GetString(ContextUserRoleKey)
}

func Error(ctx port.Context, err error) {
	ctx.JSON(shared.HTTPStatus(err), map[string]any{
		"error": map[string]any{
			"code":       shared.Code(err),
			"message":    shared.SafeMessage(err),
			"request_id": ctx.GetString(ContextRequestIDKey),
		},
	})
}

func ValidationError(ctx port.Context, message string, details any) {
	ctx.JSON(http.StatusBadRequest, map[string]any{
		"error": map[string]any{
			"code":       "validation_error",
			"message":    message,
			"details":    details,
			"request_id": ctx.GetString(ContextRequestIDKey),
		},
	})
}

func Created(ctx port.Context, payload any) {
	ctx.JSON(http.StatusCreated, payload)
}

func OK(ctx port.Context, payload any) {
	ctx.JSON(http.StatusOK, payload)
}

func NoContent(ctx port.Context) {
	ctx.Status(http.StatusNoContent)
}
