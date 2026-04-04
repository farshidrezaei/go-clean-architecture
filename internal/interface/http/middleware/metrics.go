package middleware

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Metrics struct {
	mu       sync.RWMutex
	requests map[string]uint64
	latency  map[string]float64
}

func NewMetrics() *Metrics {
	return &Metrics{
		requests: make(map[string]uint64),
		latency:  make(map[string]float64),
	}
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		key := labelsKey(c.Request.Method, route, strconv.Itoa(c.Writer.Status()))
		m.mu.Lock()
		m.requests[key]++
		m.latency[key] += time.Since(start).Seconds()
		m.mu.Unlock()
	}
}

func (m *Metrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(http.StatusOK, m.RenderPrometheus())
	}
}

func (m *Metrics) RenderPrometheus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.requests))
	for key := range m.requests {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("# HELP blog_api_http_requests_total Total number of HTTP requests handled by the API.\n")
	builder.WriteString("# TYPE blog_api_http_requests_total counter\n")
	for _, key := range keys {
		method, route, status := splitKey(key)
		builder.WriteString(fmt.Sprintf("blog_api_http_requests_total{method=%q,route=%q,status=%q} %d\n", method, route, status, m.requests[key]))
	}

	builder.WriteString("# HELP blog_api_http_request_duration_seconds_sum Cumulative request latency by route.\n")
	builder.WriteString("# TYPE blog_api_http_request_duration_seconds_sum counter\n")
	for _, key := range keys {
		method, route, status := splitKey(key)
		builder.WriteString(fmt.Sprintf("blog_api_http_request_duration_seconds_sum{method=%q,route=%q,status=%q} %.6f\n", method, route, status, m.latency[key]))
	}

	return builder.String()
}

func labelsKey(method, route, status string) string {
	return method + "|" + route + "|" + status
}

func splitKey(key string) (string, string, string) {
	parts := strings.SplitN(key, "|", 3)
	if len(parts) != 3 {
		return "unknown", "unknown", "unknown"
	}
	return parts[0], parts[1], parts[2]
}
