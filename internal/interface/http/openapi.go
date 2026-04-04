package http

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed docs/openapi.yaml docs/index.html
var docsFS embed.FS

func openAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		content, err := docsFS.ReadFile("docs/openapi.yaml")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "application/yaml", content)
	}
}

func docsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		content, err := docsFS.ReadFile("docs/index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	}
}
