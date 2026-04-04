package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	Enabled        bool
	ServiceName    string
	OTLPEndpoint   string
	OTLPInsecure   bool
	OTLPHeaders    string
	OTLPTimeout    time.Duration
	OTLPCertFile   string
	OTLPServerName string
}

func Setup(ctx context.Context, cfg Config, logger *slog.Logger) (func(context.Context) error, error) {
	if !cfg.Enabled {
		shutdown := func(context.Context) error { return nil }
		return shutdown, nil
	}

	exporter, err := newExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	logger.Info("otel tracing enabled", "service_name", cfg.ServiceName, "otlp_endpoint", cfg.OTLPEndpoint)

	return provider.Shutdown, nil
}

func newExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	if cfg.OTLPEndpoint != "" {
		options := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithTimeout(cfg.OTLPTimeout),
		}
		if headers := parseHeaders(cfg.OTLPHeaders); len(headers) > 0 {
			options = append(options, otlptracegrpc.WithHeaders(headers))
		}
		if cfg.OTLPInsecure {
			options = append(options, otlptracegrpc.WithInsecure())
		} else if creds, err := newTLSConfig(cfg); err == nil {
			options = append(options, otlptracegrpc.WithTLSCredentials(creds))
		} else {
			return nil, err
		}
		return otlptracegrpc.New(ctx, options...)
	}

	return stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stdout))
}

func newTLSConfig(cfg Config) (credentials.TransportCredentials, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if cfg.OTLPServerName != "" {
		tlsConfig.ServerName = cfg.OTLPServerName
	}
	if cfg.OTLPCertFile != "" {
		certPEM, err := os.ReadFile(cfg.OTLPCertFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(certPEM) {
			return nil, errors.New("failed to append OTLP CA certificate")
		}
		tlsConfig.RootCAs = pool
	}
	return credentials.NewTLS(tlsConfig), nil
}

func parseHeaders(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	headers := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			headers[key] = value
		}
	}
	return headers
}
