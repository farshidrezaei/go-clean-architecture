package middleware

import (
	"errors"

	"clean_architecture/internal/interface/http/common"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		spanName := c.Request.Method + " " + c.FullPath()
		if c.FullPath() == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(
			semconv.HTTPRequestMethodKey.String(c.Request.Method),
			semconv.URLPath(c.Request.URL.Path),
			attribute.Int("http.response.status_code", c.Writer.Status()),
		)
		if route := c.FullPath(); route != "" {
			span.SetAttributes(attribute.String("http.route", route))
		}
		if requestID := c.GetString(common.ContextRequestIDKey); requestID != "" {
			span.SetAttributes(attribute.String("http.request_id", requestID))
		}
		if userID := c.GetString(common.ContextUserIDKey); userID != "" {
			span.SetAttributes(attribute.String("enduser.id", userID))
		}
		if len(c.Errors) > 0 {
			span.RecordError(errors.New(c.Errors.String()))
			span.SetStatus(codes.Error, c.Errors.String())
			return
		}
		if c.Writer.Status() >= 500 {
			span.SetStatus(codes.Error, "server error")
			return
		}
		span.SetStatus(codes.Ok, "")
	}
}
