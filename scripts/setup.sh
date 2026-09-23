#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "VPSFlow Development Setup"
echo "==========================="

if ! command -v go &> /dev/null; then
    echo "ERROR: Go 1.23+ is required. Install from https://go.dev/dl/"
    exit 1
fi

echo "Go version: $(go version)"

if [ ! -f "$ROOT/.env" ]; then
    cp "$ROOT/.env.example" "$ROOT/.env"
    echo "Created .env from .env.example"
fi

MODULES=(
    "libs/go/config"
    "libs/go/errors"
    "libs/go/observability"
    "libs/go/httpx"
    "services/gateway"
)

for mod in "${MODULES[@]}"; do
    echo "Tidying $mod..."
    (cd "$ROOT/$mod" && go mod tidy)
done

echo "Running tests..."
for mod in "${MODULES[@]}"; do
    (cd "$ROOT/$mod" && go test -race -count=1 ./...)
done

echo "All tests passed!"

if command -v docker &> /dev/null && [ "${SKIP_INFRA:-}" != "1" ]; then
    echo "Starting infrastructure..."
    docker compose -f "$ROOT/platform/docker/docker-compose.yml" up -d
    echo "Infrastructure started."
fi

echo ""
echo "Setup complete! Run the gateway:"
echo "  cd services/gateway && go run ./cmd/gateway"
