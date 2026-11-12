#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

HTTP_HOST="${HTTP_HOST:-127.0.0.1}"
HTTP_PORT="${HTTP_PORT:-8080}"

if [[ -z "${JWT_SECRET:-}" ]]; then
  echo "error: JWT_SECRET is required. Add it to .env (see .env.example)." >&2
  exit 1
fi

export GIN_MODE="${GIN_MODE:-release}"

BINARY="bin/classino"

echo ""
echo "  Production: http://${HTTP_HOST}:${HTTP_PORT}"
echo "  API + static UI (single process, ${BINARY} serve)"
echo "  Config from: ${ROOT}/.env and config/config.yaml"
echo ""
echo "  Press Ctrl+C to stop."
echo ""

exec "./${BINARY}" serve
