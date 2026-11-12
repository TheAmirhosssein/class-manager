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
FRONTEND_PORT="${FRONTEND_PORT:-4173}"
API_ORIGIN="http://${HTTP_HOST}:${HTTP_PORT}"

if [[ -z "${JWT_SECRET:-}" ]]; then
  echo "error: JWT_SECRET is required. Add it to .env (see .env.example)." >&2
  exit 1
fi

pids=()

cleanup() {
  local pid
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null || true
}
trap cleanup EXIT INT TERM

start_backend() {
  if command -v air >/dev/null 2>&1; then
    air &
  else
    go run ./cmd serve &
  fi
  pids+=("$!")
}

wait_backend() {
  local i
  for i in $(seq 1 80); do
    if curl -sf "${API_ORIGIN}/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.25
  done
  echo "backend did not become ready at ${API_ORIGIN}/healthz" >&2
  return 1
}

start_frontend() {
  API_PROXY_TARGET="$API_ORIGIN" FRONTEND_PORT="$FRONTEND_PORT" node frontend/serve.cjs &
  pids+=("$!")
}

echo "Starting backend at ${API_ORIGIN}..."
start_backend
wait_backend

echo "Starting frontend at http://127.0.0.1:${FRONTEND_PORT} (API → ${API_ORIGIN})..."
start_frontend

cat <<EOF

  Backend:  ${API_ORIGIN}          (Go API + embedded static)
  Frontend: http://127.0.0.1:${FRONTEND_PORT}  (UI; /api proxied to backend)

  Set demo: false in frontend/assets/config.js for live login.
  Press Ctrl+C to stop both.

EOF

wait
