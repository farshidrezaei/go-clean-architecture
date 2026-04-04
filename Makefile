GOCACHE ?= /tmp/gocache
GOPATH ?= /tmp/gopath

.PHONY: run test test-postgres lint fmt build migrate seed truncate compose-up compose-down

run:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) go run ./cmd/api

test:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) go test ./...

test-postgres:
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) go test ./... -run Postgres

lint:
	golangci-lint run

fmt:
	gofmt -w $$(find cmd internal -name '*.go' -print)

build:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) go build ./cmd/api

migrate:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) go run ./cmd/migrate

seed:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) go run ./cmd/seed

truncate:
	GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) go run ./cmd/truncate

compose-up:
	docker compose up --build

compose-down:
	docker compose down -v
