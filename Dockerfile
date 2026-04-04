FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /blog-api ./cmd/api

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /blog-api /blog-api

EXPOSE 8080

ENTRYPOINT ["/blog-api"]
