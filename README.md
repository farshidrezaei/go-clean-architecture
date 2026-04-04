# Clean Architecture Blog API

Reference-quality Go example for a blog system API built with Clean Architecture. The project demonstrates explicit use cases, interface-based repositories, DTO mapping, Gin HTTP delivery, JWT authentication with refresh-token rotation, reuse protection, and per-device sessions, constructor injection, in-memory and PostgreSQL persistence, transaction boundaries, versioned migrations, admin APIs, request-scoped structured logging, audit-style event publishing, health checks, metrics, OpenAPI, interactive docs, and testable business rules.

## Features

- User registration, login, and refresh-token rotation
- Logout, logout-all, targeted session revocation, and refresh-token reuse protection
- JWT-protected post and comment mutations
- CRUD for posts
- CRUD for comments
- Post likes
- Role-aware authorization with `user` and `admin`
- Admin APIs for user listing and role updates
- Pagination for posts and comments
- In-memory storage for fast local runs
- PostgreSQL storage for realistic production wiring
- Structured logs with request IDs
- Audit-style event publishing through a swappable outer-layer publisher, with database append-log persistence in PostgreSQL mode
- Optional OpenTelemetry tracing with OTLP-ready config
- `/health/live`, `/health/ready`, `/metrics`, `/openapi.yaml`, and `/docs`
- Login and refresh rate limiting at the HTTP boundary, with PostgreSQL-backed persistence in PostgreSQL mode
- Docker, docker-compose, Makefile, and golangci-lint config

## Project Layout

```text
cmd/api
internal/
  domain/
    entities/
    repositories/
  usecase/
    admin/
    comment/
    post/
    shared/
    user/
  interface/
    dto/
    http/
  infrastructure/
    auth/
    config/
    database/
    health/
    logger/
    telemetry/
    repository/
```

## Why This Structure

- `domain` contains pure business models and contracts. It has no framework or database knowledge.
- `usecase` orchestrates application behavior through small interfaces, which keeps dependency inversion obvious.
- `interface` translates HTTP and JSON concerns into use-case inputs and outputs.
- `infrastructure` owns the framework, drivers, database, auth, config, and runtime details.

## Running

Copy `.env.example` to `.env` for local development if you want file-based environment loading. The application still reads standard environment variables, and `.env` is only a convenience layer for local runs.

### Memory mode

```bash
cp .env.example .env
make run
```

or

```bash
go run ./cmd/api
```

### PostgreSQL mode

```bash
cp .env.example .env
export STORAGE_DRIVER=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=blog
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_SSLMODE=disable
export JWT_SECRET='replace-me'
make migrate
make truncate
go run ./cmd/api
```

### Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

Migrations are versioned under [`migrations`](./internal/infrastructure/database/postgres/migrations) and can be applied with [`cmd/migrate`](./cmd/migrate). Seeded and runtime data can be cleared without dropping the schema via [`cmd/truncate`](./cmd/truncate).

## Example API Flow

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Farshid","email":"farshid@example.com","password":"supersecret"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"farshid@example.com","password":"supersecret"}'
```

The response contains both `access_token` and `refresh_token`.

`device_name` is optional on login and is stored with the refresh-token session.

### Refresh session

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"..."}'
```

### Logout one session

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"..."}'
```

### Logout all sessions

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout-all \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### List sessions

```bash
curl http://localhost:8080/api/v1/auth/sessions \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Revoke one session

```bash
curl -X POST http://localhost:8080/api/v1/auth/sessions/revoke \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"..."}'
```

### Create post

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Hello","content":"World","publish_now":true}'
```

### Admin: list users

```bash
curl http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

### Admin: update user role

```bash
curl -X PUT http://localhost:8080/api/v1/admin/users/$USER_ID/role \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"role":"admin"}'
```

## Operations

- Liveness: `GET /health/live`
- Readiness: `GET /health/ready`
- Metrics: `GET /metrics`
- OpenAPI: `GET /openapi.yaml`
- Interactive docs: `GET /docs`
- Tracing:
  set `OTEL_ENABLED=true`
  set `OTEL_EXPORTER_OTLP_ENDPOINT=collector:4317` for OTLP gRPC export
  optionally set `OTEL_EXPORTER_OTLP_HEADERS=authorization=Bearer token`
  optionally set `OTEL_EXPORTER_OTLP_TIMEOUT_SECONDS=5`
  optionally set `OTEL_EXPORTER_OTLP_CERT_FILE=/path/to/ca.pem`
  optionally set `OTEL_EXPORTER_OTLP_SERVER_NAME=otel-collector.internal`
  if no OTLP endpoint is set, spans fall back to stdout
- Rate limiting:
  login and refresh endpoints are protected by fixed-window limits
  memory mode uses in-process storage
  PostgreSQL mode stores windows in the database for durability across instances using the same DB
- Audit logs:
  PostgreSQL mode stores append-only events in the `audit_logs` table
  all modes still emit audit events to structured logs

## Tooling

```bash
make fmt
make test
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/blog?sslmode=disable' make test-postgres
make migrate
make seed
make truncate
make build
make compose-up
```

## GitHub Actions

The repository includes a CI workflow at [ci.yml](./.github/workflows/ci.yml), but it is intentionally manual-only for starter-kit use.

- Current trigger: `workflow_dispatch`
- It does not run on `push` or `pull_request`
- It runs format checks, `go test ./...`, and the PostgreSQL-backed integration test path
- To enable automatic CI in a derived repo, change:
  `on: workflow_dispatch`
  to:
  `on: { push: { branches: [main] }, pull_request: {} }`

When you want it to run automatically in a derived project, update the workflow triggers to include `push` and `pull_request`.

## Testing

```bash
make test
```

## Architectural Highlights

- Entity methods such as [`Post.Publish()`](./internal/domain/entities/post.go) and [`Comment.Validate()`](./internal/domain/entities/comment.go) keep business rules close to the model.
- Use cases such as [`CreatePost`](./internal/usecase/post/create_post.go) depend on interfaces, not adapters or frameworks.
- Refresh-token rotation, logout, and reuse protection are handled in [`internal/usecase/user`](./internal/usecase/user) instead of leaking session policy into handlers.
- Session records carry device metadata and can be listed or revoked individually.
- The HTTP layer performs only transport mapping and delegates all business behavior to use cases.
- Tracing is attached in middleware and bootstrap code only, so use cases remain unaware of telemetry vendors.
- Audit events are emitted from use cases through an interface and published by infrastructure sinks, keeping cross-cutting concerns out of entities.
- Brute-force protection is handled as HTTP middleware on auth endpoints, not inside use cases.
- Migrations are versioned and applied independently from the API process through [`cmd/migrate`](./cmd/migrate).
- Storage can switch between memory and PostgreSQL in [`main.go`](./cmd/api/main.go) without changing inner layers.

## Notes

- `JWT_SECRET` must be replaced in real deployments.
- Admin authorization is supported at the domain/use-case level and exposed through dedicated admin routes.
- The in-memory unit of work uses snapshotting so transaction-oriented use cases behave consistently in tests and local demos.
- The PostgreSQL repositories are intentionally explicit rather than using an ORM to keep the boundary between domain and infrastructure visible.
