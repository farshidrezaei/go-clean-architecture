# Agent Guide

## High-Signal Files

- Entry point and dependency injection: [`main.go`](./cmd/api/main.go)
- Domain entities: [`internal/domain/entities`](./internal/domain/entities)
- Repository contracts: [`internal/domain/repositories`](./internal/domain/repositories)
- User/session use cases: [`internal/usecase/user`](./internal/usecase/user)
- Post use cases: [`internal/usecase/post`](./internal/usecase/post)
- HTTP handlers and router: [`internal/interface/http`](./internal/interface/http)
- Infrastructure repositories: [`internal/infrastructure/repository`](./internal/infrastructure/repository)
- Migration files: [`internal/infrastructure/database/postgres/migrations`](./internal/infrastructure/database/postgres/migrations)
- Telemetry bootstrap: [`internal/infrastructure/telemetry`](./internal/infrastructure/telemetry)
- Audit publisher: [`internal/infrastructure/audit`](./internal/infrastructure/audit)
- Middleware stack: [`internal/interface/http/middleware`](./internal/interface/http/middleware)
- OpenAPI document: [`openapi.yaml`](./internal/interface/http/docs/openapi.yaml)

## Expected Standards

- Constructor injection only.
- Small interfaces near use-case consumption.
- Explicit error wrapping through [`shared/errors.go`](./internal/usecase/shared/errors.go).
- Both persistence implementations updated together unless the change is intentionally driver-specific.
- Session/auth changes must update JWT service, middleware expectations, DTOs, and both refresh-token repositories.
- Session-family changes must consider rotation, logout, revoke-all, and reuse detection together.
- Device-aware session changes must update DTOs, handlers, refresh-token entity fields, and both repository implementations.
- PostgreSQL schema changes should be additive and captured in a new migration file.
- Runtime concerns such as tracing, metrics, and health checks stay outside entities and use cases.
- Auth throttling belongs in middleware and should be reflected in docs/tests if thresholds or behavior change.
- PostgreSQL mode may add durable operational tables like `audit_logs` and `rate_limit_windows`; keep those changes migration-driven.
- Keep comments short and focused on rationale.

## Common Safe Refactors

- Adding new use-case input/output types
- Adding DTOs without touching entities
- Extending repository implementations to support new contracts
- Expanding tests around transaction boundaries and authorization rules
- Extending middleware for request metadata, metrics, or trace attributes
- Adding Postgres-backed integration coverage guarded by environment variables
- Updating the OpenAPI spec to match handler contracts
- Adding a new migration and extending the migrate command flow
- Extending audit events through `shared.EventPublisher` and infrastructure sinks
- Adjusting manual-only CI guidance for derived repositories

## Changes That Need Extra Care

- Altering entity methods because this can shift business semantics
- Changing repository interfaces because it can ripple through multiple use cases
- Modifying auth middleware or token contracts because handler protection depends on it
- Changing transaction behavior in either unit-of-work implementation
- Editing tracing or request-context middleware because logging, auth, and metrics all depend on the same request flow
- Changing schema for `users` or `refresh_tokens` because auth/session tests and both storage drivers depend on it
- Changing admin routes or role semantics because authorization, tests, and OpenAPI must stay aligned
- Editing existing migration files after they are considered applied
- Changing session entity fields because login, refresh, logout, session-list, and integration tests all depend on them

## Test Paths

- Fast unit/integration default: `make test`
- PostgreSQL-backed HTTP integration: `TEST_DATABASE_URL=... make test-postgres`
- Apply migrations: `make migrate`
- Local stack: `docker compose up --build`
