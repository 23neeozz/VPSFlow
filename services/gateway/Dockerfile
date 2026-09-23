# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.work ./
COPY libs/go/ libs/go/
COPY services/gateway/ services/gateway/

WORKDIR /src/services/gateway

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /gateway ./cmd/gateway

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S vpsflow && adduser -S vpsflow -G vpsflow

COPY --from=builder /gateway /usr/local/bin/gateway

USER vpsflow

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/gateway"]
