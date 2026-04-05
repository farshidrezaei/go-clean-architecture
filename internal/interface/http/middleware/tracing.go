package middleware

import (
	"errors"

	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(serviceName string) port.MiddlewareFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(ctx port.Context) {
		req := ctx.Request()
		spanName := req.Method + " " + ctx.Route()
		if spanName == req.Method+" " {
			spanName = req.Method + " " + req.URL.Path
		}

		reqCtx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header.Clone()))
		ctxReq, span := tracer.Start(reqCtx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		ctx.SetRequest(req.WithContext(ctxReq))
		ctx.Next()

		span.SetAttributes(
			semconv.HTTPRequestMethodKey.String(req.Method),
			semconv.URLPath(req.URL.Path),
			attribute.Int("http.response.status_code", ctx.ResponseStatus()),
		)
		if route := ctx.Route(); route != "" {
			span.SetAttributes(attribute.String("http.route", route))
		}
		if requestID := ctx.GetString(common.ContextRequestIDKey); requestID != "" {
			span.SetAttributes(attribute.String("http.request_id", requestID))
		}
		if userID := ctx.GetString(common.ContextUserIDKey); userID != "" {
			span.SetAttributes(attribute.String("enduser.id", userID))
		}
		if errs := ctx.Errors(); errs != "" {
			span.RecordError(errors.New(errs))
			span.SetStatus(codes.Error, errs)
			return
		}
		if ctx.ResponseStatus() >= 500 {
			span.SetStatus(codes.Error, "server error")
			return
		}
		span.SetStatus(codes.Ok, "")
	}
}
