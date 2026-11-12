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
make serve          # go run — UI + API on HTTP_PORT
make prod           # build bin/classino, GIN_MODE=release, root .env
make dev            # air/go + frontend proxy (scripts/dev.sh, root .env)
make build
air                 # hot reload (requires github.com/air-verse/air)
```

Copy `.env.example` → `.env` and set `JWT_SECRET`. Both `make dev` and `make prod` source project-root `.env`; the frontend preview (`frontend/serve.cjs`) also loads that file when run alone.
