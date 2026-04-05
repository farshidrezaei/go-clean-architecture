package ginadapter

import (
	"context"
	"net/http"

	"clean_architecture/internal/interface/http/port"
	"github.com/gin-gonic/gin"
)

type contextWrapper struct {
	ginCtx *gin.Context
}

func newContext(c *gin.Context) port.Context {
	return &contextWrapper{ginCtx: c}
}

func (c *contextWrapper) Request() *http.Request {
	return c.ginCtx.Request
}

func (c *contextWrapper) SetRequest(req *http.Request) {
	c.ginCtx.Request = req
}

func (c *contextWrapper) Context() context.Context {
	return c.ginCtx.Request.Context()
}

func (c *contextWrapper) BindJSON(v any) error {
	return c.ginCtx.ShouldBindJSON(v)
}

func (c *contextWrapper) JSON(code int, payload any) {
	c.ginCtx.JSON(code, payload)
}

func (c *contextWrapper) Status(code int) {
	c.ginCtx.Status(code)
}

func (c *contextWrapper) Param(key string) string {
	return c.ginCtx.Param(key)
}

func (c *contextWrapper) DefaultQuery(key, fallback string) string {
	return c.ginCtx.DefaultQuery(key, fallback)
}

func (c *contextWrapper) ClientIP() string {
	return c.ginCtx.ClientIP()
}

func (c *contextWrapper) GetString(key string) string {
	return c.ginCtx.GetString(key)
}

func (c *contextWrapper) Set(key string, value any) {
	c.ginCtx.Set(key, value)
}

func (c *contextWrapper) AbortWithStatusJSON(code int, obj any) {
	c.ginCtx.AbortWithStatusJSON(code, obj)
}

func (c *contextWrapper) Next() {
	c.ginCtx.Next()
}

func (c *contextWrapper) Route() string {
	return c.ginCtx.FullPath()
}

func (c *contextWrapper) ResponseStatus() int {
	return c.ginCtx.Writer.Status()
}

func (c *contextWrapper) SetResponseHeader(key, value string) {
	c.ginCtx.Writer.Header().Set(key, value)
}

func (c *contextWrapper) Errors() string {
	return c.ginCtx.Errors.String()
}

func (c *contextWrapper) Header(key string) string {
	return c.ginCtx.GetHeader(key)
}
