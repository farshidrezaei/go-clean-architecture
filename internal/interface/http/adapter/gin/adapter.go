package ginadapter

import (
	"clean_architecture/internal/interface/http/port"
	"github.com/gin-gonic/gin"
)

func Handler(fn port.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn(newContext(c))
	}
}

func Middleware(fn port.MiddlewareFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fn(newContext(c))
	}
}
