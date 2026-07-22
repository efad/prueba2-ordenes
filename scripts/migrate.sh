#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

DATABASE_URL="${DATABASE_URL:-postgres://orders:orders@localhost:5432/orders?sslmode=disable}"

if command -v go >/dev/null 2>&1; then
  go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 \
    -dir migrations postgres "$DATABASE_URL" "$@"
  exit 0
fi

docker run --rm --network host \
  -v "$ROOT_DIR/migrations:/migrations" \
  golang:1.22-alpine sh -c "
    apk add --no-cache git >/dev/null &&
    go install github.com/pressly/goose/v3/cmd/goose@v3.24.1 &&
    /go/bin/goose -dir /migrations postgres '$DATABASE_URL' $*
  "
