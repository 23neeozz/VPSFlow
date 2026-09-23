# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.work ./
COPY libs/go/ libs/go/
COPY services/auth/ services/auth/

WORKDIR /src/services/auth

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /auth ./cmd/auth

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S bosscloud && adduser -S bosscloud -G bosscloud

WORKDIR /app
COPY --from=builder /auth /usr/local/bin/auth
COPY --from=builder /src/services/auth/migrations ./migrations

USER bosscloud

ENV AUTH_MIGRATIONS_PATH=migrations
EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8081/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/auth"]
