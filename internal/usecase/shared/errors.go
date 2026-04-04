package shared

import (
	"errors"
	"fmt"

	"clean_architecture/internal/domain/entities"
)

type Error struct {
	Code    string
	Message string
	Cause   error
	Details any
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func WrapValidation(err error, message string) error {
	return &Error{Code: "validation_error", Message: message, Cause: err}
}

func WrapConflict(err error, message string) error {
	return &Error{Code: "conflict", Message: message, Cause: err}
}

func WrapNotFound(err error, message string) error {
	return &Error{Code: "not_found", Message: message, Cause: err}
}

func WrapUnauthorized(err error, message string) error {
	return &Error{Code: "unauthorized", Message: message, Cause: err}
}

func WrapForbidden(err error, message string) error {
	return &Error{Code: "forbidden", Message: message, Cause: err}
}

func WrapInternal(err error, message string) error {
	return &Error{Code: "internal_error", Message: message, Cause: err}
}

func HTTPStatus(err error) int {
	var appErr *Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "validation_error":
			return 400
		case "unauthorized":
			return 401
		case "forbidden":
			return 403
		case "not_found":
			return 404
		case "conflict":
			return 409
		}
	}

	if errors.Is(err, entities.ErrInvalidInput) || errors.Is(err, entities.ErrEmptyTitle) || errors.Is(err, entities.ErrEmptyContent) || errors.Is(err, entities.ErrEmptyCommentBody) {
		return 400
	}
	if errors.Is(err, entities.ErrUnauthorized) || errors.Is(err, entities.ErrInvalidToken) || errors.Is(err, entities.ErrTokenExpired) {
		return 401
	}
	if errors.Is(err, entities.ErrForbidden) {
		return 403
	}
	if errors.Is(err, entities.ErrNotFound) {
		return 404
	}
	if errors.Is(err, entities.ErrConflict) || errors.Is(err, entities.ErrAlreadyLiked) {
		return 409
	}

	return 500
}

func SafeMessage(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return fmt.Sprintf("request failed: %v", err)
}

func Code(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) && appErr.Code != "" {
		return appErr.Code
	}

	switch {
	case errors.Is(err, entities.ErrInvalidInput), errors.Is(err, entities.ErrEmptyTitle), errors.Is(err, entities.ErrEmptyContent), errors.Is(err, entities.ErrEmptyCommentBody), errors.Is(err, entities.ErrWeakPassword):
		return "validation_error"
	case errors.Is(err, entities.ErrUnauthorized), errors.Is(err, entities.ErrInvalidToken), errors.Is(err, entities.ErrTokenExpired):
		return "unauthorized"
	case errors.Is(err, entities.ErrForbidden):
		return "forbidden"
	case errors.Is(err, entities.ErrNotFound):
		return "not_found"
	case errors.Is(err, entities.ErrConflict), errors.Is(err, entities.ErrAlreadyLiked):
		return "conflict"
	default:
		return "internal_error"
	}
}
