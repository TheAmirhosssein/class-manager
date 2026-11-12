# class-manager (کلاسینو)

Minimal Go base: CLI, SQLite + migrations, static frontend. Full modular layout lives in [`PROJECT_ARCHITECTURE_TEMPLATE.md`](PROJECT_ARCHITECTURE_TEMPLATE.md) for when you add features.

## Layout

```text
cmd/                 CLI: serve, migrate
config/              yaml + env
frontend/            Static panel UI (index.html, assets/, docs/)
internal/
  infrastructure/
    database/        SQLite pool + Goose migrations
    server/          health routes + static frontend
  types/             shared ID type
  utils/base/        embeddable entity timestamps
pkg/errors, pkg/log
data/                SQLite file (gitignored)
```

## Commands

```sh
go mod tidy
make migrate
make serve          # http://127.0.0.1:8080 — UI + /healthz
make build
air                 # hot reload (requires github.com/air-verse/air)
```

Panel preview without Go: `node frontend/serve.cjs` (port 4173).
