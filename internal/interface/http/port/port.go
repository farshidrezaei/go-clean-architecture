package port

import (
	"context"
	"net/http"
)

type Context interface {
	Request() *http.Request
	SetRequest(*http.Request)
	Context() context.Context
	BindJSON(any) error
	JSON(int, any)
	Status(int)
	Param(string) string
	DefaultQuery(string, string) string
	ClientIP() string
	GetString(string) string
	Set(string, any)
	AbortWithStatusJSON(int, any)
	Next()
	Route() string
	ResponseStatus() int
	SetResponseHeader(string, string)
	Errors() string
	Header(string) string
}

type HandlerFunc func(Context)
type MiddlewareFunc func(Context)

type HealthChecker interface {
	Check(context.Context) error
}
