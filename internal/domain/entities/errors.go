package entities

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrAlreadyLiked      = errors.New("post already liked by user")
	ErrWeakPassword      = errors.New("password must be at least 8 characters")
	ErrEmptyTitle        = errors.New("title is required")
	ErrEmptyContent      = errors.New("content is required")
	ErrEmptyCommentBody  = errors.New("comment body is required")
	ErrPostAlreadyPublic = errors.New("post is already published")
)
