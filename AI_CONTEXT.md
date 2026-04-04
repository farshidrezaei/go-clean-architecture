# AI Context

## Purpose

This repository is a reference implementation of Clean Architecture in Go for a blog API. It now includes:

- access-token and refresh-token session flows
- refresh-token family rotation, logout, and reuse protection
- per-device session metadata, listing, and targeted revocation
- role-aware authorization (`user` and `admin`)
- in-memory and PostgreSQL persistence
- versioned PostgreSQL migrations
- request-scoped structured logging
- audit-style event publishing
- database-backed audit append log in PostgreSQL mode
- liveness/readiness/metrics endpoints
- optional OpenTelemetry tracing with OTLP-ready configuration at the HTTP/runtime boundary
- served OpenAPI document and interactive docs page
- HTTP-boundary rate limiting for sensitive auth routes
- PostgreSQL-backed durable rate-limit windows when running on PostgreSQL
- memory-backed and PostgreSQL-backed integration test paths

When making changes, preserve the dependency direction:

`domain <- usecase <- interface/infrastructure`

## Guardrails

- Do not import Gin, pgx, JWT, config, telemetry, or logging packages into `internal/domain` or `internal/usecase`.
- Keep handlers thin. They should bind requests, call a use case, and map responses.
- Put business rules in entities or use cases, not in repositories or handlers.
- Prefer adding small repository interfaces to use cases instead of widening existing ones.
- Keep repository implementations dumb and explicit. Mapping and SQL belong there; orchestration does not.
- Keep tracing, metrics, request IDs, and health checks in `internal/interface` or `internal/infrastructure`.
- If auth/session behavior changes, update both memory and PostgreSQL refresh-token repositories.
- If PostgreSQL schema changes, add a new versioned migration instead of editing old applied migrations casually.
- Keep admin-only behavior in explicit admin use cases/routes rather than branching hidden logic into generic handlers.
- Publish audit/domain events from use cases through `shared.EventPublisher`, not by importing logging directly.
- Keep brute-force and rate-limit concerns in middleware, not in use cases.

## Extension Checklist

When adding a new feature:

1. Add or evolve entities only if the business model changes.
2. Add a dedicated use case with narrow interfaces.
3. Add DTOs and a handler for transport concerns.
4. Extend both memory and PostgreSQL repositories when persistence changes.
5. Update middleware or router wiring only if the concern is transport/runtime-specific.
6. Wire the new dependency in [`main.go`](./cmd/api/main.go).
7. Add at least one test around the use case, repository, or HTTP integration behavior.

## Runtime Notes

- Tracing is optional and controlled by `OTEL_ENABLED`.
- Set `OTEL_EXPORTER_OTLP_ENDPOINT` to export spans to an OpenTelemetry Collector over gRPC.
- `OTEL_EXPORTER_OTLP_HEADERS` and `OTEL_EXPORTER_OTLP_TIMEOUT_SECONDS` tune collector export behavior.
- `OTEL_EXPORTER_OTLP_CERT_FILE` and `OTEL_EXPORTER_OTLP_SERVER_NAME` support TLS/custom CA collector setups.
- If no OTLP endpoint is configured, tracing falls back to stdout for local inspection.
- PostgreSQL integration tests run only when `TEST_DATABASE_URL` is set.
- `.env` is for local convenience; environment variables remain the source of truth.
- Database schema changes should go through `cmd/migrate` and versioned SQL files.
- `/docs` serves a Redoc page backed by `/openapi.yaml`.

## Suggested Prompts For Agents

- "Add tags to posts while preserving Clean Architecture boundaries."
- "Implement refresh tokens without leaking JWT details into use cases."
- "Add an admin moderation use case and route group."
- "Switch tracing from stdout exporter to OTLP without leaking telemetry into inner layers."
- "Add an admin-only endpoint and update both auth and integration tests."
- "Add a new migration and update both Postgres tests and OpenAPI."
