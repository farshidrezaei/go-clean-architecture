package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv            string
	HTTPPort          string
	JWTSecret         string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration
	StorageDriver     string
	DatabaseURL       string
	LogLevel          string
	OTELEnabled       bool
	OTELServiceName   string
	OTLPTraceEndpoint string
	OTLPInsecure      bool
	OTLPHeaders       string
	OTLPTimeout       time.Duration
	OTLPCertFile      string
	OTLPServerName    string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration
}

func Load() Config {
	databaseURL := buildDatabaseURL()
	return Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessTTL:      getDuration("JWT_ACCESS_TTL_MINUTES", 60),
		JWTRefreshTTL:     getDuration("JWT_REFRESH_TTL_MINUTES", 24*7*60),
		StorageDriver:     getEnv("STORAGE_DRIVER", "memory"),
		DatabaseURL:       databaseURL,
		LogLevel:          getEnv("LOG_LEVEL", "INFO"),
		OTELEnabled:       getBool("OTEL_ENABLED", false),
		OTELServiceName:   getEnv("OTEL_SERVICE_NAME", "blog-api"),
		OTLPTraceEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTLPInsecure:      getBool("OTEL_EXPORTER_OTLP_INSECURE", true),
		OTLPHeaders:       getEnv("OTEL_EXPORTER_OTLP_HEADERS", ""),
		OTLPTimeout:       getDurationSeconds("OTEL_EXPORTER_OTLP_TIMEOUT_SECONDS", 5),
		OTLPCertFile:      getEnv("OTEL_EXPORTER_OTLP_CERT_FILE", ""),
		OTLPServerName:    getEnv("OTEL_EXPORTER_OTLP_SERVER_NAME", ""),
		ReadTimeout:       getDurationSeconds("HTTP_READ_TIMEOUT_SECONDS", 10),
		WriteTimeout:      getDurationSeconds("HTTP_WRITE_TIMEOUT_SECONDS", 10),
		ShutdownTimeout:   getDurationSeconds("HTTP_SHUTDOWN_TIMEOUT_SECONDS", 10),
	}
}

func (c Config) Validate() error {
	switch c.StorageDriver {
	case "memory", "postgres":
	default:
		return fmt.Errorf("unsupported STORAGE_DRIVER %q", c.StorageDriver)
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.JWTAccessTTL <= 0 || c.JWTRefreshTTL <= 0 {
		return errors.New("JWT TTL values must be positive")
	}
	if c.StorageDriver == "postgres" && c.DatabaseURL == "" {
		return errors.New("database configuration is required when STORAGE_DRIVER=postgres")
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getDuration(key string, fallbackMinutes int) time.Duration {
	raw := getEnv(key, strconv.Itoa(fallbackMinutes))
	value, err := strconv.Atoi(raw)
	if err != nil {
		value = fallbackMinutes
	}
	return time.Duration(value) * time.Minute
}

func getDurationSeconds(key string, fallbackSeconds int) time.Duration {
	raw := getEnv(key, strconv.Itoa(fallbackSeconds))
	value, err := strconv.Atoi(raw)
	if err != nil {
		value = fallbackSeconds
	}
	return time.Duration(value) * time.Second
}

func getBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func buildDatabaseURL() string {
	if direct := os.Getenv("DATABASE_URL"); direct != "" {
		return direct
	}

	host := getFirstNonEmpty("DB_HOST", "POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := getFirstNonEmpty("DB_PORT", "POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	name := getFirstNonEmpty("DB_NAME", "POSTGRES_DB")
	if name == "" {
		name = "blog"
	}
	user := getFirstNonEmpty("DB_USER", "POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	password := getFirstNonEmpty("DB_PASSWORD", "POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	sslMode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		name,
		url.QueryEscape(sslMode),
	)
}

func getFirstNonEmpty(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
