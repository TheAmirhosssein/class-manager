# Project Architecture Template

This template describes a reusable architecture for a single-deployable Go service. It is a modular monolith: one module, one binary, several long-running processes, and a vertical slice per business module. Inside each slice the flow is transport → use case → repository, with technical adapters behind small interfaces.

It is not a microservice system, and it is not strict hexagonal architecture. Repository interfaces live next to their implementations. Use cases depend on those interfaces. The process root is the only place that chooses concrete adapters.

User-facing copy, domain nouns, and external provider names from any one product do not belong in a new project. Replace every `<placeholder>` with the new project's names.

Stack assumed by this template:

- Go module, standard layout (`cmd`, `internal`, `pkg`)
- CLI entrypoint (Cobra or equivalent)
- HTTP router with middleware
- Server-rendered HTML templates plus JSON for mutations
- PostgreSQL through a small query API (`pgx` pool or equivalent), no ORM
- SQL migrations (Goose-style up/down files)
- Redis for sessions, short-lived records, and list-backed queues
- Object storage for uploaded files
- Cookie sessions backed by Redis
- Password hashing behind an interface
- Structured logging
- Integration tests against real Postgres (and real object storage when needed) started with Docker, plus an in-memory Redis for tests

---



## 1. Architecture Overview



### Overall architecture

One binary serves three roles, started by a `serve` command and stopped together on `SIGINT` / `SIGTERM`:

1. **HTTP server** — HTML pages, JSON mutations, and WebSocket upgrades.
2. **Queue consumer** — reads named lists from Redis and dispatches module consumers.
3. **Optional workers** — long-running clients (for example an external WebSocket feed) that call a module use case. Workers are gated by configuration and exit when disabled by waiting on context cancellation.

Operational commands (`populate`, privileged-user creation, and similar) live on the same CLI. They build only the dependencies they need and then exit.

The UI is server-rendered. A shared layout receives one response envelope. Mutations are submitted by static JavaScript as JSON and answered with the same envelope. Live updates travel through Redis lists into in-process channels and out through WebSocket broadcasts.

### Main architectural style

**Modular monolith with vertical slices and a shallow layered core.**

Each business area is a package tree under `internal/apps/<module>`. A slice owns its entities, enums, result messages, input params, repositories, use cases, HTTP delivery, and queue consumers. Shared technical code stays in `internal/infrastructure`, `internal/httpconfig`, `internal/types`, `internal/utils`, and `pkg`.

A **composition module** (the shell / panel) is allowed to read across modules so every HTML page can render the same chrome. It does not own the writes of those modules.

### Major layers


| Layer          | Where it lives                                                            | Role                                                                                         |
| -------------- | ------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Process        | `cmd`, `internal/infrastructure/server`, `internal/infrastructure/worker` | Start and stop processes, construct adapters, register module routers and consumers          |
| Transport      | `internal/apps/<module>/delivery`, `internal/httpconfig`, `frontend`      | HTTP, WebSocket, templates, request models, route constants, auth middleware                 |
| Application    | `internal/apps/<module>/usecase`                                          | One operation per method: validate, authorize, call repositories and ports, return a message |
| Module model   | `entity`, `enum`, `validation`, `params`, `message`                       | Data the use case and repository share, closed sets of values, pure rules, result codes      |
| Persistence    | `internal/apps/<module>/repository`                                       | SQL or Redis access for that module                                                          |
| Infrastructure | `internal/infrastructure/<adapter>`                                       | Database pool, cache, queue, hasher, notifier, object storage                                |
| Shared kernel  | `internal/types`, `internal/utils`, `pkg`                                 | IDs, nullables, pagination, result type, logging, errors                                     |




### Dependency direction

Dependencies point inward and sideways only through explicit seams:

```text
cmd
  → infrastructure/server | infrastructure/worker | seed
      → module delivery (routers, consumers)
          → handler → usecase → repository interface
                             → infrastructure port interface
                             → entity, enum, params, message
          → repository implementation → database driver | redis
frontend templates and static files have no Go imports
pkg does not import internal
```

The process root may import every module's delivery package. A module use case may import another module's **entity, enum, or repository interface** when it coordinates that data. A module must not import another module's delivery package.

### Separation of concerns

- Handlers translate HTTP into use-case inputs and use-case messages into HTML or JSON. They do not decide business rules and they do not write SQL.
- Use cases decide whether an operation is allowed and what happens next, including whether to enqueue a follow-up message. They do not know about `gin`, templates, or SQL text.
- Repositories load and store rows. A missing row is an empty value, not a business error. The use case decides that "empty" means "not found".
- Infrastructure adapters wrap one external system and expose a small interface.
- Route strings, the response envelope, the template renderer, and generic auth middleware live in one transport kernel so handlers and templates share them.
- User-facing success and failure text lives in the module `message` package, not in handlers.



### Main design principles

1. **One module, one reason to change.** A feature's entity, use case, repository, handler, and template move together.
2. **Constructors wire dependencies.** `NewX` receives interfaces and config. Globals are limited to process singletons created at startup (config, database pool, cache client, object-storage client).
3. **Interfaces sit beside the implementation that fulfills them.** Callers in other packages depend on the interface. A compile-time assertion (`var _ Interface = (*impl)(nil)`) keeps them aligned.
4. **Operations return a message, not a bare error.** Business failures are values with a stable code and an HTTP class. Unexpected failures are logged and replaced with a generic server message.
5. **The database is the source of truth.** Unique constraints enforce identity. Redis holds sessions, expiring secrets, and queues. Queues are for live fan-out and asynchronous follow-up work, not for the primary write.
6. **Money and measured quantities use decimals.** Do not store them as binary floats.
7. **Time is stored in UTC** and converted to the user timezone at the presentation edge.
8. **Authorization is checked twice for protected writes:** once at the route (middleware) and again in the use case against the persisted actor.
9. **Tests exercise real SQL.** Use cases and repositories run against a migrated Postgres. Pure functions get ordinary unit tests.
10. **Do not add a layer that has only one implementation and no boundary.** An interface is justified for a repository, an external system, or a test double. It is not justified for a one-off helper.

---



## 2. Directory Structure

```text
<project>/
├── cmd/
│   └── main.go                         # CLI: serve, populate, operational commands
├── config/
│   ├── config.go                       # Typed config, yaml + environment, load-once
│   └── config.yaml                     # Non-secret structure and defaults
├── frontend/
│   ├── templates/
│   │   ├── base/                       # Layout, list chrome, shared partials
│   │   ├── errors/                     # Not found, forbidden, server error
│   │   ├── <module>/                   # One HTML file per page
│   │   └── landing/                    # Public pages, if any
│   └── static/
│       ├── css/
│       ├── js/
│       │   ├── request/                # JSON mutations, one file per form flow
│       │   └── <feature>.js            # Page behavior, WebSocket clients
│       ├── fonts/
│       ├── images/
│       └── sound/
├── internal/
│   ├── apps/
│   │   ├── <module>/
│   │   │   ├── entity/
│   │   │   ├── enum/
│   │   │   ├── message/
│   │   │   ├── params/
│   │   │   ├── repository/
│   │   │   ├── usecase/
│   │   │   ├── validation/             # Only when a rule is pure and reused
│   │   │   ├── <capability>/           # Optional module-local capability (PDF, pricing)
│   │   │   └── delivery/
│   │   │       ├── http/
│   │   │       │   ├── handler/
│   │   │       │   │   └── model/
│   │   │       │   └── router/
│   │   │       └── consumer/
│   │   └── <composition-module>/       # Shell / read model used by HTML pages
│   │       ├── entity/
│   │       ├── message/
│   │       ├── model/                  # View models for the shell
│   │       ├── repository/
│   │       ├── usecase/
│   │       └── <composition-module>.go # Renders the shared layout around a page
│   ├── httpconfig/                     # Route catalog, envelope, renderer, auth middleware
│   ├── infrastructure/
│   │   ├── cache/
│   │   ├── database/
│   │   │   └── migrations/
│   │   ├── filemanager/
│   │   ├── hash/
│   │   ├── queue/
│   │   ├── sender/
│   │   ├── server/
│   │   └── worker/
│   ├── seed/                           # Named fixtures for tests and the populate command
│   ├── types/                          # ID and nullable scalars
│   └── utils/
│       ├── base/                       # Entity, pagination, process-wide router bag
│       ├── messages/                   # The result type every use case returns
│       └── testutil/
├── pkg/
│   ├── errors/
│   ├── log/
│   ├── testdocker/                     # Test-only containers
│   ├── timezone/
│   └── validator/                      # Pure input checks with no domain import
├── nginx/
│   └── default.conf
├── scripts/                            # One-off operational scripts, not the request path
├── .air.toml                           # Dev rebuild: go and html, runs `serve`
├── .env.example                        # Names only, placeholder values
├── docker-compose.dev.yaml            # Database, cache, object storage
├── docker-compose.prod.yaml           # Reverse proxy, app, database, cache, object storage
├── Dockerfile.Prod
├── Makefile
├── go.mod
└── go.sum
```



### `cmd/`

Responsibility: process entry. Parse the CLI, start the server, consumer, and workers, or run a short operational command.

Belongs here: the root command, `serve`, `populate`, and commands that an operator runs by hand.

Must not contain: business rules, SQL, route tables, or HTTP handlers. A command may call one use case after constructing its repositories, the same way the router does. Keep that wiring short.

### `config/`

Responsibility: load configuration once and expose a typed struct.

Belongs here: `config.yaml` for names, ports, queue names, pagination, template paths, session key names, and bucket names. Environment bindings for secrets and environment-specific URLs.

Must not contain: credentials, production tokens, or business decisions.

### `frontend/templates/`

Responsibility: server-rendered pages. `base/` is the layout every authenticated page extends. `errors/` are the pages the transport kernel renders for 403, 404, and 500. `<module>/` holds one template per screen, named after the page (`<resource>-list.html`, `<resource>-profile.html`).

Belongs here: markup, template variables that match the response envelope (`Message`, `Panel`, `Routes`, `Data`, `AdditionalData`), and shared partials such as pagination.

Must not contain: business calculations that the use case should have done, or copies of route strings. Use the route values passed in the envelope or template filters that call the route resolvers.

### `frontend/static/`

Responsibility: assets the HTTP server exposes at `/static`.

Belongs here: CSS, fonts, images, page scripts, and `js/request/` modules that `fetch` JSON to a form's `action`.

Must not contain: secrets, generated dependency trees, or a second copy of rules that the use case enforces. Client checks are only for immediate feedback.

`frontend/src` is not part of the runtime. The server serves `templates` and `static` directly. Add a frontend build only when a new project truly needs one, and commit or generate the files the server actually serves.

### `internal/apps/<module>/`

Responsibility: one business slice.

Belongs here: everything listed in the tree above for that slice.

Must not contain: the database pool setup, the HTTP server, another module's handlers, or shared route constants (those live in `httpconfig`).

Important subdirectories:


| Directory                      | Responsibility                                          | Belongs                                                       | Does not belong              |
| ------------------------------ | ------------------------------------------------------- | ------------------------------------------------------------- | ---------------------------- |
| `entity/`                      | In-memory records, list wrappers, queue payload structs | Fields, small methods that answer questions about one value   | SQL, HTTP, repository calls  |
| `enum/`                        | Closed string sets stored in the database               | Constants, `IsValid`, display helpers                         | I/O                          |
| `message/`                     | Success and failure results for this module             | Codes, user-facing text, exported message values              | Formatting of HTTP responses |
| `params/`                      | Input structs                                           | Use-case inputs and repository query inputs, kept separate    | JSON tags for HTTP bodies    |
| `repository/`                  | Persistence                                             | Interface, SQL or Redis implementation, row scanning          | Authorization, HTTP codes    |
| `usecase/`                     | Application operations                                  | Orchestration, authorization, calls to ports                  | `gin.Context`, SQL strings   |
| `validation/`                  | Pure rules reused by a use case                         | Functions of values in, boolean or error out                  | Database access              |
| `<capability>/`                | A document or calculation that is not a use case        | Rendering a PDF, a pricing function owned by the module       | A second public API          |
| `delivery/http/handler/`       | HTTP adapters                                           | Bind, call use case, choose template or JSON                  | Business branching           |
| `delivery/http/handler/model/` | Request bodies and page view models                     | `json`/`binding` tags, mapping from entities                  | Repository calls             |
| `delivery/http/router/`        | Module composition                                      | Construct repos and use cases, register routes and middleware | Business rules               |
| `delivery/consumer/`           | Queue handlers                                          | Unmarshal a payload, call a use case or forward to a channel  | Owning the poll loop         |




### `internal/apps/<composition-module>/`

Responsibility: the shared application shell. It loads the current actor and the small cross-module summaries the layout needs, then renders the page template with the standard envelope.

Belongs here: read models and read-only SQL that join other modules' tables for display. The composition type is passed into handlers that render HTML.

Must not contain: the writes of those other modules, and must not become a place for new business rules. If a rule decides whether a write is allowed, it belongs in the owning use case.

### `internal/httpconfig/`

Responsibility: the transport kernel shared by every module.

Belongs here: the route catalog and path resolvers, the response envelope, pagination mapping, template renderer and template filters, HTML error pages, query parsing (`page`, `page-size`), multipart-to-file mapping, and generic middleware (`guest`, `authenticated`, `admin`, `end-user`).

Must not contain: SQL, module use-case construction, or policy that belongs to one resource. A middleware that needs a use case is registered by that module's router and receives the use case as an argument. Do not let this package construct repositories.

### `internal/infrastructure/`

Responsibility: technical systems and process runners.


| Directory              | Responsibility                     | Belongs                                                             | Does not belong                                                        |
| ---------------------- | ---------------------------------- | ------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| `cache/`               | Redis client lifetime              | Parse URL, return a shared client                                   | Business keys and TTLs (those belong to the repository that uses them) |
| `database/`            | Postgres pool and migration runner | Pool settings, embed of `migrations/`, test database setup          | Queries for a module                                                   |
| `database/migrations/` | Schema history                     | Goose up/down SQL                                                   | Application code                                                       |
| `filemanager/`         | Object storage                     | `Save` interface, vendor adapter, client setup, test container hook | Decisions about who may upload                                         |
| `hash/`                | Password hashing                   | `Encrypt` / `Compare` interface and implementation                  | When a password change is allowed                                      |
| `queue/`               | List queue and consumer loop       | `Push`, consumer registration, blocking read                        | Payload meaning                                                        |
| `sender/`              | Outbound notifications             | Interface, live client, no-op or mock used in tests                 | When to notify                                                         |
| `server/`              | HTTP server and consumer process   | Listen, graceful shutdown, call each module `Register`              | Feature logic                                                          |
| `worker/`              | Long-running external clients      | Reconnect, parse the foreign message, call one use case             | HTML or SQL                                                            |




### `internal/seed/`

Responsibility: a known dataset. `CreateSeed` fills Postgres and Redis. Tests and the `populate` command both call it.

Belongs here: exported fixture values and insert functions, ordered so foreign keys succeed.

Must not contain: production secrets, or data that tests depend on without a name. Tests refer to `seed.<Fixture>`, not to magic strings copied in each file.

### `internal/types/`

Responsibility: the scalar types every module shares.

Belongs here: `ID`, `CacheID`, nullable wrappers that implement SQL scanning, and constructors (`NewNullString`, `NewNullID`, and so on). `ID.Valid()` is false for the zero value.

Must not contain: entities or business rules.

### `internal/utils/`


| Directory   | Responsibility                                   | Belongs                                                                                                                                                                | Does not belong             |
| ----------- | ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------- |
| `base/`     | Types embedded or reused everywhere              | `Entity` (`ID`, `CreatedAt`, `UpdatedAt`), `CacheEntity`, `Pagination`, `FileSave`, the `Router` bag (engine, config, db, redis, websocket hub, object-storage client) | Feature rules               |
| `messages/` | The result type                                  | `Message` with description, code, HTTP code, and optional error                                                                                                        | Module-specific text        |
| `testutil/` | Assertions that do not belong in production code | Time comparison and similar helpers                                                                                                                                    | Fixtures (those are `seed`) |


Query-fragment helpers that know about SQL placeholders belong next to the database, under `internal/infrastructure/database` or `internal/utils` — not in `pkg`. They must bind user values as parameters. Column names passed into them are developer-controlled literals.

### `pkg/`

Responsibility: libraries that do not know about a module.

Belongs here: the coded error type, the logger, timezone conversion, locale digit formatting, pure validators, and Docker helpers used only by tests.

Must not contain: imports of `internal`, HTTP handlers, or SQL for a feature. If a helper needs `internal/types` or a database driver, it is not a `pkg` library.

### `nginx/`, `scripts/`, repo root

`nginx/` is the production reverse proxy: the app upstream, long timeouts for WebSockets, and a storage prefix that proxies object storage. Do not put application routes here that the Go server already serves.

`scripts/` holds operator shell scripts that are not part of a request. Prefer a CLI command when the task needs the application's code.

Root files wire the developer loop: `Makefile`, Air config, compose files, Dockerfile, `.env.example`, `go.mod`.

---



## 3. Layer Responsibilities


| Layer                                                   | Responsibility                                               | Can depend on                                                                                                                                                        | Must not depend on                                                             |
| ------------------------------------------------------- | ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Process (`cmd`, `server`, `worker`)                     | Start, wire, shut down                                       | Config, infrastructure clients, module routers and consumers, module use cases (workers and CLI only)                                                                | Business rules inlined in `main`                                               |
| Transport (`delivery/http`, `httpconfig`, templates)    | Routes, binding, status codes, HTML and JSON shape, sessions | Use cases, handler models, entities for mapping, composition module, config                                                                                          | Repository implementations, SQL, another module's `delivery`                   |
| Application (`usecase`)                                 | One operation: authorize, validate, read, write, enqueue     | Own entity, enum, params, message, validation; repository interfaces; infrastructure port interfaces; another module's repository, entity, or enum when coordinating | `gin`, templates, handler models, SQL text, concrete driver types in new ports |
| Module model (`entity`, `enum`, `validation`)           | Values and pure rules                                        | `types`, `utils/base`, another module's `enum` or `entity` only for a real association                                                                               | Repositories, use cases, HTTP, drivers                                         |
| Result catalog (`message`)                              | Stable codes and user-facing text                            | `utils/messages`, `pkg/errors`                                                                                                                                       | Handlers, repositories                                                         |
| Persistence (`repository`)                              | Queries and cache records                                    | Own entity and params, `types`, driver, log, SQL helpers                                                                                                             | Use cases, handlers, the HTTP response type                                    |
| Infrastructure adapter                                  | One external system                                          | Config, the vendor SDK, log                                                                                                                                          | Module use cases, except `server` and `worker`                                 |
| Composition module                                      | Read model for the shell                                     | Other modules' repositories, entities, and enums for reads                                                                                                           | Their delivery packages; their write rules                                     |
| Shared kernel (`types`, `utils/messages`, `utils/base`) | Cross-module language                                        | Standard library and decimal type                                                                                                                                    | A specific module                                                              |
| `pkg`                                                   | Reusable technical code                                      | Standard library and third-party libraries                                                                                                                           | `internal`                                                                     |
| Frontend                                                | Presentation                                                 | The route and envelope the server rendered                                                                                                                           | A private API the server does not already expose                               |


The database driver type appears in repository constructors (`*pgxpool.Pool`, `*redis.Client`) because the interface and the implementation share a package. That is accepted. Use cases must depend on the repository interface, not on the pool. New ports that cross a package boundary (`Hash`, `Queue`, `SmsSender`, `FileManager`) are interfaces so a use case does not import the vendor SDK.

---



## 4. Package Responsibilities



### `cmd`

Why it exists: one `main` package for the binary.

Owns: command definitions and signal handling. `serve` cancels a shared context and waits for the HTTP server, the consumer, and workers.

Exposes: nothing importable.

Allowed dependencies: `config`, `infrastructure/server`, `infrastructure/worker`, `infrastructure/database`, `infrastructure/cache`, `seed`, and the specific module packages an operational command calls.

Never place here: a route, a query, or a rule that a use case already owns.

### `config`

Why it exists: one load path for yaml and environment.

Owns: the `Config` struct, nested sections, `GetConfig()` (load once, panic on failure), a partial `GetTestConfig()` for tests that only need a bucket name or similar, and `InTestMode()` based on `-test.` arguments.

Exposes: `GetConfig() *Config`.

Allowed dependencies: the env/yaml loader and the standard library.

Never place here: a connection pool or a client. Config describes the system; infrastructure opens it.

### `internal/apps/<module>/entity`

Why it exists: the records use cases and repositories pass around.

Owns: the struct embedded with `base.Entity`, list types that include `base.Pagination`, queue payload structs with JSON tags, and methods that answer a question about that value (`Admin()`, commission lookup).

Exposes: exported types. Construction is struct literals, not a framework.

Allowed dependencies: `types`, `utils/base`, decimal, and another module's `entity` or `enum` when the association is real.

Never place here: `json` request bodies (those are handler models), SQL tags as a second model, or I/O.

### `internal/apps/<module>/enum`

Why it exists: a closed set of strings that match a database enum.

Owns: an unexported struct created once, exported fields (`"ACTIVE"` style values), `IsValid`, and translation or rank helpers.

Exposes: a package-level variable such as `<Name>` and functions like `<Name>Translate`.

Allowed dependencies: standard library only, unless a helper truly needs nothing else.

Never place here: database queries to "discover" values. The Go constants and the migration enum are updated together.

### `internal/apps/<module>/message`

Why it exists: every outcome of the module has one definition.

Owns: two files, `message.go` (success) and `errors.go` (failures). Each outcome is a package-level `messages.Message`. Success values use `NewMessage`. Failures use `NewMessageByError` so `Error` is set and `HttpCode()` is derived from the numeric code.

Exposes: exported variables (`<Entity>Created`, `<Entity>NotFound`).

Allowed dependencies: `internal/utils/messages`, `pkg/errors`.

Never place here: sentences assembled in handlers, or codes reused with a different meaning.

### `internal/apps/<module>/params`

Why it exists: keep HTTP models, use-case inputs, and SQL inputs from collapsing into one struct.

Owns: `<Entity>Usecase<input>Params` (page, page size, filters, actor identity) and `<Entity>Repo<input>Params` (limit, offset, the same filters). The use case converts page to offset.

Exposes: exported structs.

Allowed dependencies: `types` and other entities only when a param carries an already-loaded actor.

Never place here: `binding` tags or template formatting.

### `internal/apps/<module>/repository`

Why it exists: the only SQL and the only Redis key layout for the module.

Owns: the exported interface, the unexported struct, `New<Entity>Repository`, and `var _ Interface = (*impl)(nil)`. Methods are `Create`, `Read`, `ReadBy<Field>`, `Update`, `Delete`, `ExistBy<Field>`, plus the specific queries the use case needs. List methods return `([]Entity, totalCount, error)`.

Exposes: the interface and the constructor. The struct stays private.

Allowed dependencies: entity, params, types, the pool or Redis client, log, SQL helpers.

Never place here: HTTP status, user-facing messages, or "if the role is admin" decisions. A missing row returns the zero entity and `nil` when the driver reports no rows. Any other driver error is returned.

Cache-backed repositories (short-lived secrets) use the same shape: `Create`, `Get`, `Delete`, `Exist`, with TTL stored on the entity.

### `internal/apps/<module>/usecase`

Why it exists: the application API of the module.

Owns: a struct of interfaces, `New<Entity>Usecase`, and one exported method per operation. Methods take `context.Context` and either an entity plus scalars or a params struct. They return `(result, messages.Message)` or only `messages.Message`.

Exposes: the concrete use-case type and constructor. Handlers depend on this type. There is no extra use-case interface unless a second implementation exists.

Allowed dependencies: own entity, enum, message, params, validation; own repository interfaces; infrastructure interfaces; another module's repository interface, entity, or enum; log; config when the operation needs a setting (TTL, bucket, feature flag).

Never place here: SQL, template names, session writes, or `gin.Context`. On a driver error, log it with operation context and return the shared server message. On a rule failure, return the module message and do not log it as an internal error.

### `internal/apps/<module>/validation`

Why it exists: a rule that is awkward inline and must be unit-tested without a database.

Owns: pure functions.

Never place here: a rule that is a single comparison only the use case needs. That stays in the use case. Do not create this directory until the second caller or the test needs it.

### `internal/apps/<module>/delivery/http/handler`

Why it exists: the HTTP adapter.

Owns: a handler struct holding use cases, config, the composition module, and the WebSocket hub when the feature streams. Methods bind input, read the actor from the request context, call the use case, and write the response.

Exposes: `New<Entity>Handler` and the handler methods the router registers.

Allowed dependencies: usecase, handler/model, entity and enum for mapping, httpconfig, composition module, config.

Never place here: a repository field, a query, or a copy of a rule the use case enforces.

Page reads call the composition module's render method with a template path and a `httpconfig.Response`. JSON mutations write `ctx.JSON(message.HttpCode, response)`. Binding failures are a bad-request message and stop the handler.

### `internal/apps/<module>/delivery/http/handler/model`

Why it exists: the HTTP shape is not the entity.

Owns: request structs with `json` and `binding` tags (camelCase JSON) and view models built by `New<Entity>Model` functions. View models format dates, translated enums, and display signs. List view models embed `httpconfig.Pagination`.

Never place here: a password hash, a write, or a field the template does not need. Do not pass the raw entity to a template when the view needs formatting.

### `internal/apps/<module>/delivery/http/router`

Why it exists: the module's composition root.

Owns: a router struct that holds the shared `*base.Router`, the composition module, and repository interfaces. `New<Module>Router` constructs repositories from `router.DB` and `router.Redis`. `Register` attaches each resource. Resource methods construct use cases and handlers and register routes on `router.Server` using `httpconfig` constants and middleware.

Exposes: `New<Module>Router` and `Register`.

Allowed dependencies: this module's handler, usecase, repository; other modules' repository constructors and use cases only when this module's screen truly orchestrates them; infrastructure constructors (`NewHash`, `NewSmsSender`, `NewQueue`, `NewMinioAdaptor`); httpconfig.

Never place here: business conditions. Middleware order is part of this package: session store first, then guest or auth, then role, then resource-specific gates.

One module router may call sub-registrations as methods (`authRouter`, `<resource>Router`) so the file stays split per resource without a second public type.

### `internal/apps/<module>/delivery/consumer`

Why it exists: turn a queue payload into a use case call or an in-process broadcast.

Owns: a consumer type with `Register`, which binds queue names from config to handlers. Handlers unmarshal into the entity payload struct. Follow-up work calls a use case. Live UI updates send on a package-level channel that the WebSocket handler reads.

The package name is `consumer` (singular) for every module.

Never place here: the poll loop. That stays in `infrastructure/queue`. Never import another module's `delivery` package. Reading another module's repository or enum is allowed when the follow-up belongs to this module.

### `internal/apps/<composition-module>`

Why it exists: every HTML page needs the same actor card, navigation counts, and asset summary.

Owns: read entities, a repository of cross-table reads, a use case that returns messages, view models, and one method `PanelResponse(ctx, status, template, response)` that fills `response.Panel` and calls `ctx.HTML`. Status codes at or above the error range are delegated to `httpconfig` error pages instead of the feature template.

Other modules' handlers may depend on this type. This module's use case must not call other modules' use cases in a cycle. It reads through repositories.

### `internal/httpconfig`

Why it exists: one catalog for paths and one shape for every response.

Owns:

- Constants `/<module>/<resource>/<action>/` and `/ws/<module>/<resource>/<event>`, plus `XResolver` functions that substitute path parameters.
- `Response` with `Message`, `Panel`, `Routes`, `Data`, `AdditionalData`.
- `NewResponse(message, data)` which copies the message code and description and fills `Routes` from the catalog.
- Template renderer that maps a `Response` onto template context and registers resolver filters.
- `AuthRequired`, `GuestOnly`, `AdminOnly`, and the end-user counterpart. They read the session keys from config.
- Helpers `GetPage` and `GetPageSize`.

Allowed dependencies: config, the HTTP framework, `utils/base`, `utils/messages`, the composition module's **view model** for the envelope, and role constants. Resource policy that needs a database is injected from the owning router.

### `internal/infrastructure/database`

Why it exists: one pool and one migration path.

Owns: pool limits, driver type registration (decimal), `GetDb`, embedded migrations, `SetupTestDB` which starts Postgres, migrates, and returns the pool. The `Makefile` applies the same SQL directory with the migration CLI for local development.

Never place here: a table's queries.

### `internal/infrastructure/queue`

Why it exists: modules publish bytes without knowing the consumer loop.

Owns: `Queue` with `Push(ctx, []byte) error`, constructed with a Redis client and a list name. `QueueConsumer` holds the client, the pool (so module consumers can build repositories), and `Register(name, handler)`. `Run(ctx)` blocks reading each registered list and exits when the context is cancelled.

`Push` is a list push. Do not add a second pub/sub path for the same queue. Live delivery inside the process uses Go channels after the consumer has unmarshaled the payload.

### `internal/infrastructure/hash`, `sender`, `filemanager`, `cache`

Each owns one interface and its implementation:


| Package       | Interface                                            | Implementation                                            |
| ------------- | ---------------------------------------------------- | --------------------------------------------------------- |
| `hash`        | `Encrypt`, `Compare`                                 | bcrypt                                                    |
| `sender`      | one method per notification kind                     | HTTP client to the provider, plus a mock that returns nil |
| `filemanager` | `Save(ctx, bucket, name, contentType, reader, size)` | object-storage adapter that creates the bucket if needed  |
| `cache`       | none; returns `*redis.Client`                        | shared client from `REDIS_URL`                            |


Use cases depend on `Hash`, `SmsSender`, `FileManager`, and `Queue`. They do not depend on `*redis.Client` except repositories that are themselves the cache adapter (constructed by the router).

### `internal/infrastructure/server`

Why it exists: the HTTP and consumer processes are not a module.

Owns: `Run(ctx, conf)` and `RunConsumer(ctx, conf)`. `Run` builds the engine, static file route, template renderer, pool, Redis, WebSocket hub, and `base.Router`, then calls each module `NewXRouter(...).Register()`, then shuts the listener down with a short timeout when `ctx` is done. `RunConsumer` builds one `QueueConsumer` and calls each module consumer `Register` before `Run`.

This is the only production package that should import every module router.

### `internal/infrastructure/worker`

Why it exists: a process that speaks a foreign protocol.

Owns: reconnect with backoff (start at one second, cap at thirty, reset after a good connection), respect context cancellation, translate the foreign payload, and call one use case method. The worker is constructed with the same repositories and queues a router would use.

Never place here: HTML, or a second copy of the pricing or validation rules. Those stay in the use case or `validation`.

### `internal/seed`

Why it exists: tests and `populate` share one dataset.

Owns: exported fixture entities, counts, default page size, and `CreateSeed(ctx, db, redis)`. Postgres inserts run in foreign-key order. Redis fixtures are separate functions called from `CreateSeed`.

### `internal/utils/messages` and `pkg/errors`

`pkg/errors` owns `CustomError` with a numeric code, `NewError`, `Error2Custom`, and `HttpCode()`. `HttpCode` returns the first three digits, so `404_104` is HTTP 404. Shared constructors: `NewServerError` (`500_00`) and `NewBadRequestError` (`400_00`).

`utils/messages` owns `Message` and constructors `NewMessage`, `NewMessageByError`, `NewServerErrorMessage`, `NewBadRequestErrorMessage`.

### `pkg/log`

Owns three loggers (`Info`, `Warning`, `Error`) on one JSON handler with source locations, writing to the process log file, and sets the default logger. Feature code uses these loggers rather than inventing another logger.

### `pkg/testdocker`

Owns starting and removing disposable Postgres and object-storage containers, waiting until they accept connections, and generating container names. Used by `SetupTestDB` and file-storage tests. Not used by production code.

### `pkg/timezone` and `pkg/validator`

`timezone` owns `NowUTC` and conversion into the product timezone, plus locale-calendar formatting used by view models. `validator` owns pure checks (for example a phone-number pattern) that more than one module may call. A check that only one entity understands stays in that module.

---



## 5. Dependency Rules



### Who may import whom


| From                             | May import                                                                                                                                                              | Must not import                                |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| `cmd`                            | `config`, `infrastructure/server`, `worker`, `database`, `cache`, `seed`, a module use case for a CLI command                                                           | `frontend`, handler models                     |
| `infrastructure/server`          | every module's `delivery` router or consumer, `httpconfig`, `database`, `cache`, `utils/base`                                                                           | A module `usecase` body copied inline          |
| `infrastructure/worker`          | the one module `usecase` and `repository` it feeds, `queue`, `cache`, `database`                                                                                        | Handlers, templates                            |
| `<module>/delivery/http/router`  | own handler, usecase, repository; infrastructure constructors; `httpconfig`; composition module; another module's repository or usecase when the screen orchestrates it | Another module's `delivery`                    |
| `<module>/delivery/http/handler` | own usecase, model, entity, `httpconfig`, composition module                                                                                                            | `repository`, SQL driver                       |
| `<module>/usecase`               | own entity, enum, params, message, validation, repository interface; infrastructure interfaces; another module's repository, entity, enum                               | `delivery`, `gin`, SQL                         |
| `<module>/repository`            | own entity, params, types, driver                                                                                                                                       | usecase, delivery, another module's repository |
| `<module>/entity`                | `types`, `utils/base`, another module's `enum` or `entity` for an association                                                                                           | repository, usecase, delivery                  |
| `<module>/message`               | `utils/messages`, `pkg/errors`                                                                                                                                          | entity I/O, HTTP                               |
| `httpconfig`                     | config, HTTP framework, messages, base, role enum, composition **view model**                                                                                           | Module repositories                            |
| `pkg`                            | standard library, third-party                                                                                                                                           | `internal`                                     |
| any `internal`                   | `pkg`                                                                                                                                                                   | —                                              |
| composition module               | other modules' entity, enum, repository                                                                                                                                 | other modules' delivery or usecase             |
| `seed`                           | entities, enums                                                                                                                                                         | delivery, usecase                              |




### Where interfaces live


| Interface                     | Package                      | Implemented by                     |
| ----------------------------- | ---------------------------- | ---------------------------------- |
| `<Entity>Repository`          | `<module>/repository`        | private struct in the same file    |
| `Queue`                       | `infrastructure/queue`       | private struct in the same package |
| `Hash`                        | `infrastructure/hash`        | bcrypt struct                      |
| `SmsSender` (or `<Notifier>`) | `infrastructure/sender`      | live client and mock               |
| `FileManager`                 | `infrastructure/filemanager` | object-storage adapter             |


There is no `domain/ports` package. Adding one duplicates the interfaces that already sit with their adapters.

### Where implementations live

Concrete types stay unexported in the package that owns the interface. Other packages see `NewX(...) Interface`. Use cases are concrete types, not interfaces. Handlers are concrete types.

### Dependency inversion

The use case asks for `repository.AccountRepository` or `hash.Hash`. The router, worker, or CLI command calls `NewAccountRepository(pool)` and passes the interface. Tests pass the same constructors pointed at the test pool or the mock sender.

Do not generate a mock for every repository when the test can run against Postgres. Use the mock sender and in-memory Redis because those are the ports with a side effect outside the database.

### How circular dependencies are avoided

- Delivery packages are imported only by `infrastructure/server` and by their own module.
- Module A may call module B's **repository** while module B calls module A's **repository**. That is a module-level cycle and it is tolerated for coordination. It must not become a **package import** cycle. `entity` must not import `usecase`. `repository` must not import `usecase`. `usecase` must not import `delivery`.
- If two entities need each other, keep one side as an `ID` instead of embedding the other entity, or move the shared enum to a package that neither entity's parent imports in a cycle.
- The composition module is a leaf that others call for rendering. It does not call them back through HTTP types.
- `httpconfig` must not import a use case. If a middleware needs one, the router passes it in.
- `pkg` must not import `internal`, or `internal` and `pkg` will cycle the moment `internal` imports that `pkg` subpackage.



### Process-wide singletons

`GetConfig`, `GetDb`, `GetClient` (Redis), and `GetMinIO` are created once. Feature code receives the resulting values through constructors and `base.Router`. New code must not call `GetConfig()` from a repository or use case. Pass the `*config.Config` or the one field you need.

---



## 6. Domain Structure

This project does not use a separate domain-service layer, aggregate root framework, value-object types, command bus, or domain-event bus. The concepts below are the ones that actually carry the design.

### Entities

An entity is a struct in `<module>/entity` embedding `base.Entity`:

```text
ID        types.ID
CreatedAt time.Time
UpdatedAt time.Time
```

Associations are either an embedded related entity (when the query always loads it) or an `ID` / `NullID`. List screens use `<Entity>List` containing the slice plus `base.Pagination`.

Queue payloads are separate structs in the same package, with JSON tags, named for the fact (`<Entity>Create`, `<Entity>Update`). They are not the row entity. Producers marshal them. Consumers unmarshal them.

Cache-backed records embed `base.CacheEntity` (`CacheID` and `Duration`) instead of `base.Entity`.

Entities may contain methods that only read their fields. They do not load themselves.

### Enums

Stored as strings and as database enum types. The Go value and the SQL value are the same literal. `IsValid` is the check use cases run before writing. Display text is a function in the enum package or in the handler model, not a column.

### Validation

Structural validation of JSON uses binding tags on handler models. Semantic validation (ranges, cross-field rules, "this status may follow that status") lives in the use case, or in `<module>/validation` when the rule is pure. Shared format checks (phone numbers) live in `pkg/validator`.

### Repositories

One interface per aggregate or cache record. The use case is the unit of work at the method level: it may call several repositories. There is no generic repository and no hidden ORM session.

### Use cases

A use case type groups the operations of one aggregate or one workflow (`Auth`, `<Entity>`, a document workflow). Methods are verbs: `Create`, `Read`, `Update`, `Delete`, `LoginWithPassword`, `ConfirmOTP`. They are not split into command and query types. A read method still returns a message so the handler can choose the status code.

Authorization that depends on the actor and the target is a private method on the use case, used by every write that needs it.

### Params

Two structs when the shapes differ. The handler fills the use-case param from the query string and the session. The use case builds the repository param, including `Limit` and `Offset`.

### Results and errors

There is no `domain/errors` package. Failures are `messages.Message` values defined in `<module>/message`. A failure has `Error != nil`. A success has `Error == nil`. Codes are integers written with a digit separator: the first three digits are the HTTP class, the rest are unique inside the service.

```text
<httpStatus>_<moduleBlock><sequence>

400_00     shared bad request
500_00     shared internal error
400_1xx    module A validation
404_1xx    module A missing
409_1xx    module A conflict
403_1xx    module A forbidden
422_1xx    module A semantic rejection
200_1xx    module A read success
201_1xx    module A created
202_1xx    module A accepted update
```

Assign each module a block (`1xx`, `3xx`, …) and do not reuse a code.

### Events

There is no domain-event type. Asynchronous facts are the queue payload structs. Publishing is an explicit `json.Marshal` plus `queue.Push` inside the use case that decided the fact happened. Consuming is explicit in `delivery/consumer`.

### Module-local capabilities

A package such as a statement or PDF renderer stays inside the module, takes entities in, and returns bytes or a view struct. The handler calls it after the use case has loaded the data. It does not open its own database connection.

### Composition read models

The shell has its own entity types even when the columns originate elsewhere. Mapping into those types happens in the composition repository, so the layout does not depend on every field of every module entity.

---



## 7. Infrastructure



### Database

`infrastructure/database` opens one `pgxpool` from `POSTGRES_URL`, with an explicit max and min connection count, a health-check period, and decimal type registration. Module repositories receive the pool. They use parameterized SQL (`$1`, `$2`, …) and scan into entities.

List queries that need a total use one statement that returns rows and the count (a `WITH data AS (...), rows_count AS (...)` query or an equivalent), not two round trips hidden in a framework.

Search and optional filters go through a helper that appends bound arguments and returns an `AND` fragment. The helper never concatenates the user's string into the SQL text.

### Repositories

See section 4. A Postgres repository and a Redis repository for the same module both live in `repository/`. The constructor makes the backing store obvious: `NewXRepository(db)` or `NewXRepository(redis)`.

### External HTTP and WebSocket APIs

Outbound HTTP lives in the adapter that owns the protocol (`sender`, or a dedicated client package under `infrastructure/<provider>` if the integration is larger than a notification). The use case sees a method such as `SendOTP(mobile, code)`.

A long-lived inbound socket lives in `infrastructure/worker`. It reconnects with backoff and calls a use case. The foreign JSON struct stays next to the worker, not in the module entity, unless the field names are already the module's language. Map foreign names into use-case arguments at the worker boundary.

### Message broker

Redis lists are the broker. Names come from config (`<purpose>_queue`). Producers depend on `queue.Queue`. The consumer process is the only subscriber. There is no topic exchange and no outbox table.

### Caches

Redis is used in three ways, and they stay separate:


| Use                                       | Owner                                                  | Key                                            |
| ----------------------------------------- | ------------------------------------------------------ | ---------------------------------------------- |
| Session cookies                           | HTTP session middleware, configured in the auth router | the session library's keys                     |
| Short-lived secrets                       | the module repository                                  | a prefix plus the subject, TTL on the record   |
| Queues                                    | `infrastructure/queue`                                 | the configured list name                       |
| Optional read cache for an external quote | the client that fetches it                             | private to that client, with a TTL from config |


Do not store the primary entity in Redis as a cache-aside layer unless a measured case needs it. The database remains authoritative.

### Authentication

Sessions are created only by the auth handler after the use case returns success. The session stores the account id, the login identifier, and the role, under key names from config. The store is Redis. `AuthRequired` copies those values into the request context. Handlers read the actor from the context and pass the identifier into the use case. The use case loads the actor from the database when the decision depends on current state, not only on the session role.

Password checks go through `hash.Hash`. The repository stores the hash bytes. Clear-text passwords exist only in the request model and the use-case argument.

### Configuration

Described in section 10. Infrastructure reads it through the struct it was given, not through `os.Getenv` scattered in adapters. The one exception is process startup inside `GetConfig`.

### Logging

`pkg/log` is the logger. Log unexpected errors with a stable message and fields (`error`, and the id or key involved). Do not log passwords, session secrets, tokens, or one-time codes.

### Metrics and traces

The current design has no metrics registry and no application-level tracer. Do not invent a parallel observability stack inside a feature. If a project adds them, the HTTP server and the worker own the instrumentation, and module code receives a narrow interface or uses the standard library's `log/slog` fields already in `pkg/log`. Vendor metrics middleware stays in `infrastructure/server`.

### Third-party integrations

Each vendor gets an interface in `infrastructure/<concern>` (`sender`, `filemanager`, or `<provider>`). The router chooses the live implementation or the mock from configuration (`Active` flags). Disabled integrations return success without calling the network, when the product requirement is "skip", or they no-op the whole worker.

### How adapters connect

```text
router or worker or CLI
    constructs repository with the pool or Redis
    constructs Queue / Hash / Sender / FileManager with config
    passes interfaces into NewXUsecase
use case calls the interface
adapter talks to the vendor
```

---



## 8. API / Transport Layer



### Style

The service is a website with JSON side doors, not a public REST platform and not gRPC.

- **Reads of pages** are `GET` and return HTML.
- **Mutations** are `POST`, `PUT`, or `DELETE` and return JSON.
- **Live updates** are `GET` WebSocket routes under `/ws/...`.
- Paths end with `/` except the WebSocket routes, which match the upgrade handler's path.
- Path parameters use the router's `:name` form.
- One resource lives under one prefix: `/<module>/<resource>/<action>/`. Do not repeat the module name in the resource segment.

Examples of the shape, not of a specific product:

```text
GET    /<module>/<resource>/list/
POST   /<module>/<resource>/create/
PUT    /<module>/<resource>/update/:id/
DELETE /<module>/<resource>/delete/:id/
GET    /<module>/<resource>/:id/
GET    /ws/<module>/<resource>/updates
```

Every path constant and every resolver lives in `httpconfig`. Handlers, templates, and JavaScript `action` attributes use those constants (rendered into the page). A new path is added in one place.

### Routes

Module `Register` attaches routes. The server does not use a global annotation scanner. Middleware is applied per route or per group explicitly so a public page cannot inherit `AuthRequired` by accident.

### Handlers

A handler method:

1. Reads the actor from context when the route is authenticated.
2. Binds the body into a handler model. On failure, responds with the bad-request message and returns.
3. Maps the model to an entity or a use-case param. Parses `:id` with `httpconfig.ParseParamToID`.
4. Calls exactly the use case that owns the operation.
5. If `message.Error != nil` and the screen is HTML, lets the composition renderer or `HandleErrorMessage` select the error page.
6. If the screen is JSON, writes `NewResponse(message, data)` with `message.HttpCode`.
7. If the screen is HTML and the message is success, builds the view model and calls `panel.PanelResponse`.

Session creation and clearing stay in the auth handler. Other handlers do not write the session.

### Request and response models

Request models live in `handler/model` and describe the JSON the browser sends. Response data for JSON is either `gin.H` for a handful of fields or a view model. The envelope is always:

```text
message.code
message.description
data
additionalData
panel          # HTML only; empty in pure JSON
routes         # catalog of links for the layout
```

JavaScript reads `message.description` and treats HTTP 2xx as success.

### Validation

Binding tags enforce presence and simple formats before the use case runs. The use case repeats semantic checks and returns a module message. Do not trust the client. The handler does not translate validator internals into a new error vocabulary; a bind failure is the shared bad-request message.

### Middleware

Applied in this order when they apply:

1. Session middleware, installed once, before authenticated routes.
2. `GuestOnly` for login pages, `AuthRequired` for the application.
3. `AdminOnly` or the end-user gate for role-specific pages.
4. A resource gate registered by the owning router (for example "this actor already has a pending request") when the product needs it. The gate receives a use case; it does not construct repositories itself.

`AuthRequired` redirects anonymous browser users to the login path. `GuestOnly` redirects an existing session away from login. Role failure renders the forbidden page.

### Authentication and authorization

Authentication answers "who is this session". Authorization answers "may this role perform this action", in middleware for the route class and in the use case for the resource. A use case that only checks the session role and never the loaded record is incomplete for any action where roles can change or the target is not the caller.

### Error handling at the edge


| Situation                       | HTTP                                    | Body                                                                   |
| ------------------------------- | --------------------------------------- | ---------------------------------------------------------------------- |
| Bind failure                    | 400                                     | bad-request message                                                    |
| Module message with `Error` set | `message.HttpCode`                      | that message, HTML error page when the handler is rendering a document |
| Module success                  | `message.HttpCode` (200, 201, 202, 204) | envelope, or redirect when the flow is a classic form login            |
| Unexpected panic                | framework default                       | do not add a second recovery that hides the log                        |


500 responses use the generic server page or the generic server message. They do not include the driver error text.

### Serialization

JSON encoding is the HTTP framework's. HTML encoding is the template engine (Django-like blocks and `{{ }}`). The renderer accepts the `Response` struct and exposes its fields to the template. Dates shown to people are formatted in the view model or by a template filter that calls `pkg/timezone`. The database value stays UTC.

Static files are served from the configured static directory at `/static`. Template files are loaded from the configured template directory. Both paths are resolved from the process working directory's project root.

### Request flow

```text
browser
  → reverse proxy (production)
    → gin engine
      → session middleware
        → auth and role middleware
          → handler
            → bind model, map to entity or params
              → use case
                → repository (SQL or Redis)
                → port (hash, sender, queue, file storage)
                → module message
            ← message and optional result
          ← HTML via composition renderer, or JSON envelope
```

WebSocket flow:

```text
use case Push(payload)
  → Redis list
    → consumer process LPOPs / BLPOP
      → module consumer unmarshals
        → use case for follow-up work
        → or an in-process channel
          → HTTP server's WebSocket handler broadcasts to connected browsers
```

The HTTP process and the consumer process do not share memory. The channel exists only inside the process that needs to push to sockets it holds. A consumer that must change the database calls a use case. A consumer that must update open browsers publishes to the channel the socket handler reads. If those are different processes, the socket handler is the one subscribed to that queue, in the HTTP process. Do not assume a channel crosses processes.

---



## 9. Database & Migrations



### Organization

One PostgreSQL database. Tables are plural snake_case (`accounts`, `<resources>`). Columns are snake_case. Primary keys are `id` identity / `SERIAL` and map to `types.ID` (`int64`). Foreign keys are `<singular>_id`. Timestamps are `created_at` and `updated_at`, defaulting to the database clock, stored as timestamps without a local offset and treated as UTC by the application.

Closed sets that the application branches on are PostgreSQL enum types named `enum_<table>_<column>` with the same literals as the Go enum.

Amounts that can be fractional use `NUMERIC`, scanned into `decimal.Decimal`. Integer money uses `BIGINT` only when the unit is already the minor unit and cannot be fractional. Do not add `FLOAT` or `DOUBLE PRECISION` for quantities.

Optional columns use SQL `NULL` and a `types.Null*` field. `NewNull*` treats the zero value as not valid, which matches "absent". Do not use a sentinel string for null.

### Schema ownership

The migration directory owns the schema. Repositories own the queries. A new column is added in a migration first, then mapped in the repository scan list and the entity. Both directions of a migration (`Up` and `Down`) are required.

Indexes and unique constraints that express identity (`UNIQUE` mobile number, unique external selector) live in SQL. The use case still checks existence so it can return a conflict message, and it still handles the unique-violation race as an internal error or a conflict, not as a panic.

### Migration location and naming

```text
internal/infrastructure/database/migrations/<YYYYMMDDHHMMSS>_<verb>_<object>.sql
```

Verbs: `create`, `add`, `change`, `drop`. Examples of the shape: `create_<table>_table`, `add_<column>_to_<table>`.

Files use Goose markers:

```text
-- +goose Up
-- +goose StatementBegin
...
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
...
-- +goose StatementEnd
```

The directory is embedded in the test database setup so tests migrate themselves. Operators apply the same directory with the `Makefile` target against `POSTGRES_URL`. Do not keep a second migration tree.

### Transactions

Each repository method is one statement and one implicit transaction. The code base does not open a multi-statement `BEGIN`. A use case that writes and then enqueues is **not** atomic with the queue.

Rules for new work:

- Prefer a single SQL statement.
- If two writes must succeed or fail together, put them in one repository method that begins and commits a transaction, or in a repository method that accepts a transaction the caller owns. Do not start putting SQL in the use case to achieve this.
- Do not enqueue a follow-up before the write has succeeded. If the follow-up is required for correctness, the consumer must tolerate a retry, and the write that records "this follow-up is done" must be idempotent (unique key on the source id).
- There is no unit-of-work type. Do not add one until a real multi-aggregate transaction needs it.



### Repository versus database package


| Concern                                    | Package                                    |
| ------------------------------------------ | ------------------------------------------ |
| Pool, migrations, test database            | `infrastructure/database`                  |
| A table's SQL                              | that module's `repository`                 |
| Cross-table read for the shell only        | composition module repository              |
| Placeholder fragments and "no rows" checks | SQL helper next to the database, not `pkg` |


The "no rows" check maps the driver's no-rows error to `(zero, nil)`. Callers use `entity.ID.Valid()` to mean "found".

---



## 10. Configuration



### Structure

`config.Config` is a tree of sections: application name and version, session, pagination, template paths, constants (timeouts, query-parameter names), HTTP host and port, database URL, Redis URL and queue names, each external provider, object storage, and feature flags.

Yaml tags mark file fields. Env tags mark environment fields. `env-required:"true"` fails startup when a value is missing.

### Where values live


| Kind                             | Location                        | Examples of names                                                                                                             |
| -------------------------------- | ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Structure and safe defaults      | `config/config.yaml`            | HTTP port, page size, queue names, session key names, template paths, bucket name, OTP lifetime                               |
| Secrets and per-environment URLs | environment, via `.env` locally | database URL, Redis URL, session secret, provider API key, object-storage credentials, external service token, feature on/off |
| Computed                         | code                            | project root, derived from the working directory                                                                              |


`.env` is loaded only when the file exists. `.env` is gitignored. `.env.example` lists every variable with placeholders and is committed. Never commit a real `.env`, and never paste real tokens into yaml.

`config.yaml` in this design also contains no secrets. A value that would hurt if the repository leaked belongs in the environment even if a sample yaml once had it.

### Defaults and validation

Required fields have no silent default: missing config panics at startup with the load error. Numeric defaults that are product choices (page size, OTP lifetime, cache TTL) live in yaml so they are visible. `GetTestConfig` returns the smallest struct a unit test needs so pure tests do not read the developer's `.env`.

`InTestMode` lets adapters avoid production side effects when `go test` is the process. Prefer injecting a mock into the use case over branching on `InTestMode` inside business logic.

### How configuration reaches the application

```text
GetConfig() once at process start
  → server.Run(ctx, conf)
  → worker.Run(ctx, conf)
  → base.NewRouter(..., conf, ...)
      → module router stores router.Config
          → use case constructor receives *config.Config or the fields it needs
```

CLI commands call `GetConfig` or open the pool through `GetDb`, which reads config. They do not parse flags for secrets that already have environment variables.

### Feature flags

A provider that must not run in development has an `Active` boolean from the environment. The adapter or worker checks it at the edge and skips the network. The use case still records the local state.

---



## 11. Error Handling



### Philosophy

Expected outcomes are data. Unexpected outcomes are logs plus a generic server message. Callers branch on `message.Error != nil` and on the specific message value in tests, not on string contains.

### Domain and application errors

Defined in `<module>/message/errors.go` as `messages.Message` values. They cover validation, not found, conflict, forbidden, and semantic rejection. The numeric code's first three digits are the HTTP status. The description is safe to show to the user.

### Infrastructure errors

Driver errors, HTTP client errors, and filesystem errors stay as `error` until the use case boundary. The repository logs the query failure and returns the error. The use case logs it again only if the repository did not already log that same failure. Pick one layer and keep it: **log at the repository for query failures, and log at the use case for failures of ports the use case called directly** (hash, sender, queue, file storage). Do not log the same error in both places and also in the handler.

The handler does not log a business message. It may log a failure to save the session, because that is a transport concern.

### Transport errors

Bind errors become the shared bad-request message. Status selection is `message.HttpCode`. HTML error pages exist for forbidden, not found, and server error. Other 4xx responses on a JSON route return the envelope. A redirect is used for login success, logout, and anonymous access to a private page.

### Wrapping

`pkg/errors.CustomError` is the wrap format for business failures. Do not wrap a business failure in a second code. Infrastructure errors may be wrapped with `%w` inside an adapter when the adapter adds context; the use case still does not return that string to the client.

`Error2Custom` on a non-custom error becomes a generic 500. That is why use cases must convert unknown errors into `NewServerErrorMessage()` themselves before they reach the handler. The handler should not receive a raw driver error.

### Mapping

```text
repository error == no rows     → zero entity, nil
repository other error          → log, return error
use case sees error             → log if this layer owns it, return server message
use case sees invalid rule      → return module message, Error set
use case succeeds               → return module success message, Error nil
handler sees Error != nil       → HTTP message.HttpCode, no driver text
handler sees bind error         → 400 bad-request message
```



### Logging

JSON, with source, to the process log file. Fields are lowercase snake case. The HTTP framework's request log is the access log. There is no separate audit log. Security-relevant successes (login) are visible as normal application flow; do not log the password or the one-time code.

### Retryable versus not


| Error                                                 | Retry                                                                                                                                                 |
| ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Validation, forbidden, not found, conflict            | No. The client must change the request.                                                                                                               |
| 500 after a local bug or a bad query                  | No automatic retry on the HTTP request.                                                                                                               |
| WebSocket disconnect in a worker                      | Yes. Exponential backoff from 1s to 30s, stop on context cancel, reset after a successful connection.                                                 |
| Notification provider HTTP 5xx                        | The current sender does not retry. A new sender may retry only idempotent sends, with a small cap, inside the adapter.                                |
| Queue payload that does not unmarshal                 | Not retryable. Log it and drop it. Fix the producer.                                                                                                  |
| Queue handler that fails after the payload was popped | Treat as retryable work that can be lost. New handlers of required work must be idempotent and should only acknowledge after success. See section 16. |


---



## 12. Testing Strategy



### What is tested where


| Layer                                                                 | Test type                     | How                                                                                                                |
| --------------------------------------------------------------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `validation`, pricing, timezone, SQL-helper fragments, enum `IsValid` | Unit                          | Table-driven, no Docker                                                                                            |
| `usecase`                                                             | Integration                   | Real Postgres with migrations and `seed`, in-memory Redis, mock sender, real hash                                  |
| `repository`                                                          | Integration                   | Same database setup, assert rows through the repository's own reads                                                |
| `hash`                                                                | Unit                          | Library behavior, no containers                                                                                    |
| `handler`                                                             | Not the default               | Add one only when mapping or status selection is subtle. Do not duplicate every use-case case at the HTTP layer.   |
| Browser end-to-end                                                    | Not part of the default suite | The template engine and static JS are verified by running the app. There is no Selenium suite in the architecture. |
| `seed`                                                                | Indirectly                    | Every integration `TestMain` loads it. A broken fixture fails many packages.                                       |




### Layout and naming

Tests sit next to the code: `<file>_test.go`. The test package is the external package (`package usecase_test`, `package repository_test`) so tests use only exported API.

Test function names: `Test<Type>_<Method>`. Cases are a slice of structs with `name` and the expected `messages.Message` or `wantErr`. Subtests use `t.Run` and `t.Parallel` when they do not mutate the same fixture row. Cases that update a shared seeded row must not run in parallel with each other.

Use `testify/require`. Compare business failures with `require.EqualError` against the expected message's error. Compare successes with `require.NoError(t, message.Error)` and then read back through the repository.

### Integration harness

Each package `TestMain`:

1. Starts a Postgres container and migrates it (`database.SetupTestDB`).
2. Starts in-memory Redis.
3. Calls `seed.CreateSeed`.
4. Runs tests.
5. Removes the container.

Tests require a working Docker daemon. That is intentional. Do not replace the repository tests with a hand-written fake that reimplements SQL.

Object-storage tests start a disposable object-storage container the same way, in the packages that save files.

### Mocks and fixtures

- Fixtures are `internal/seed`, not JSON files scattered through tests.
- The notification port has `NewMockSender` in the sender package.
- Redis in tests is in-memory, not a container, because the session and OTP behavior under test is the client protocol, not Redis durability.
- Do not mock the repository in use-case tests. The use-case test is allowed to touch the database; that is the point.
- `internal/utils/testutil` holds comparisons that `require` does not express well (clock fields).



### What not to test

Do not assert log text. Do not test the HTTP framework. Do not copy a migration into a test as `CREATE TABLE`. Do not depend on test order except for the shared seed.

### Make target

`make test` runs the packages that contain tests. New packages follow the same `TestMain` pattern so one command covers them. Pure unit packages do not need a container.

---



## 13. Naming Conventions


| Thing                     | Convention                                                    | Example shape                                                 |
| ------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- |
| Module directory          | lower singular noun                                           | `internal/apps/<module>`                                      |
| Package                   | lower, no underscores, matches the folder                     | `usecase`, `repository`, `httpconfig`                         |
| Go file                   | lower, no underscores, `<noun><role>.go`                      | `<entity>usecase.go`, `<entity>repo.go`, `<entity>handler.go` |
| Test file                 | `<file>_test.go`                                              | `<entity>usecase_test.go`                                     |
| Exported type             | PascalCase                                                    | `<Entity>Usecase`, `<Entity>Repository`                       |
| Unexported implementation | camelCase                                                     | `<entity>Repo`                                                |
| Constructor               | `New` + type                                                  | `New<Entity>Repository`                                       |
| Interface                 | role noun, no `I` prefix                                      | `<Entity>Repository`, `Hash`, `Queue`                         |
| Use-case method           | verb, CRUD spelled `Create` `Read` `Update` `Delete`          | `ReadByMobileNumber`                                          |
| Repository method         | same verbs, `ExistBy<Field>` for booleans                     | `ExistByName`                                                 |
| Handler method            | the screen or action                                          | `List`, `Create`, `Profile`                                   |
| Message variable          | `<Entity><Outcome>`                                           | `<Entity>Created`, `<Entity>NotFound`                         |
| Error code constant       | unexported `code<Outcome>`                                    | `code<Entity>NotFound`                                        |
| Params                    | `<Entity>Usecase<Action>Params`, `<Entity>Repo<Action>Params` |                                                               |
| View model                | `<Entity>Model`, constructor `New<Entity>Model`               |                                                               |
| Request model             | `<Action><Entity>Model`                                       | `Create<Entity>Model`                                         |
| Enum Go value             | PascalCase field, SCREAMING_SNAKE string                      | `Active: "ACTIVE"`                                            |
| Enum SQL type             | `enum_<table>_<column>`                                       |                                                               |
| Context parameter         | `ctx`                                                         |                                                               |
| Error variable            | `err`                                                         |                                                               |
| Receiver                  | short, one letter or one syllable                             | `r` repository, `u` use case, `h` handler                     |
| Config field              | PascalCase matching a yaml or env name                        | `PageSize`                                                    |
| Environment variable      | SCREAMING_SNAKE                                               | `<SERVICE>_URL`, `<SERVICE>_API_KEY`                          |
| Database table            | plural snake_case                                             | `<entities>`                                                  |
| Database column           | snake_case                                                    | `created_at`, `<entity>_id`                                   |
| Migration file            | `<timestamp>_<verb>_<object>.sql`                             |                                                               |
| HTTP path constant        | `<Module><Resource><Action>`                                  |                                                               |
| HTTP path                 | `/<module>/<resource>/<action>/`                              | trailing slash                                                |
| WebSocket path            | `/ws/<module>/<resource>/<event>`                             | no requirement to end in slash                                |
| JSON field                | camelCase                                                     | `firstName`                                                   |
| Template file             | `<module>/<page>.html`                                        |                                                               |
| Template blocks           | the layout's block names (`title`, `style`, content)          |                                                               |
| Static request script     | `js/request/<feature>-request.js`                             |                                                               |
| Queue config key          | `<purpose>_queue`                                             | snake_case value is the Redis list name                       |
| Queue payload             | `<Entity><Fact>`                                              | `<Entity>Create`                                              |
| Redis cache key           | `<module>_<purpose>_<subject>`                                | built in the use case or repository, not in handlers          |
| Seed variable             | exported noun                                                 | `<Entity><FixtureName>`                                       |
| Test                      | `Test<Type>_<Method>`                                         |                                                               |
| CLI command               | lower verb                                                    | `serve`, `populate`                                           |


Package comments and identifiers are English. User-facing description strings may be the product language and stay inside `message`.

File names do not use underscores (`accounttypeusecase.go`, not `account_type_usecase.go`). Abbreviations already used in a module stay consistent (`repo` in file names, `Repository` in type names).

---



## 14. Dependency & External Service Patterns



### Relational database


|                |                                                                  |
| -------------- | ---------------------------------------------------------------- |
| Interface      | `<module>/repository`, method set of the aggregate               |
| Implementation | same package, `*pgxpool.Pool`                                    |
| Wiring         | module router or worker calls `New<Entity>Repository(router.DB)` |
| Configuration  | `POSTGRES_URL`                                                   |
| Errors         | no-rows → empty value; anything else → log and return            |
| Tests          | `SetupTestDB` plus seed                                          |




### Redis cache and sessions


|                |                                                                                             |
| -------------- | ------------------------------------------------------------------------------------------- |
| Interface      | repository interface for data the module owns; session library for cookies                  |
| Implementation | `*redis.Client` from `infrastructure/cache`                                                 |
| Wiring         | `GetClient()` at process start, passed through `base.Router`                                |
| Configuration  | `REDIS_URL`, session URL, session secret, session key names                                 |
| Errors         | missing key on `Exist` is false, not a 500; other client errors are infrastructure failures |
| Tests          | in-memory Redis                                                                             |




### Queue


|                |                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------- |
| Interface      | `queue.Queue` (`Push`)                                                                        |
| Implementation | Redis list, key from config                                                                   |
| Consumer       | `QueueConsumer.Register` from `<module>/delivery/consumer`                                    |
| Wiring         | router builds a `Queue` per list the use case publishes; server builds one consumer           |
| Errors         | push failure is a server message after the local write decision; see section 16 for ack rules |
| Tests          | in-memory Redis, assert the list content or run the handler function directly                 |




### Object storage


|                |                                                                                                             |
| -------------- | ----------------------------------------------------------------------------------------------------------- |
| Interface      | `FileManager.Save`                                                                                          |
| Implementation | vendor adapter, bucket created on first save                                                                |
| Wiring         | `GetMinIO` on the router bag, `NewMinioAdaptor` in the router                                               |
| Configuration  | endpoint, access key, secret, bucket name                                                                   |
| Errors         | failure to store is a server message; the use case does not write the database row if the file did not save |
| Tests          | disposable container, or not involved when the use case is given a fake `FileManager`                       |


Public-read bucket policy is only for assets the browser must fetch anonymously. Private files use a private bucket and an authorized URL. Do not copy a public policy onto a bucket that holds documents.

### Notification provider


|                        |                                                                                       |
| ---------------------- | ------------------------------------------------------------------------------------- |
| Interface              | `SmsSender` or a narrower notifier interface                                          |
| Implementation         | HTTP client, plus `NewMock...` for tests                                              |
| Configuration          | base URL, API key, sender identity, template id, `Active`                             |
| Errors                 | non-success HTTP status becomes an error; the use case turns it into a server message |
| Behavior when inactive | return nil without sending, when the flag means "suppressed"                          |
| Tests                  | mock, never the real provider                                                         |




### External streaming feed


|                |                                                                                   |
| -------------- | --------------------------------------------------------------------------------- |
| Interface      | none beyond the use case method the worker calls                                  |
| Implementation | `infrastructure/worker`                                                           |
| Configuration  | URL, credential, `Active`                                                         |
| Errors         | dial and read failures log and reconnect; malformed messages log and continue     |
| Tests          | unit-test the parser if it is non-trivial; do not open a live socket in `go test` |




### Password hashing


|                |                                                                                    |
| -------------- | ---------------------------------------------------------------------------------- |
| Interface      | `Hash`                                                                             |
| Implementation | bcrypt with the library default cost                                               |
| Errors         | hash failure is a server message; mismatch is an authentication message, not a 500 |


---



## 15. Background Jobs / Workers

The `serve` command starts:

```text
context cancelled on SIGINT or SIGTERM
    ├── HTTP server        server.Run
    ├── queue consumer     server.RunConsumer
    └── workers            worker.Run<Name>   (one goroutine each)
wait group until each returns
```



### HTTP server lifecycle

Bind the configured host and port. On context cancellation, shut down with a timeout of a few seconds so in-flight page renders finish. Log a listen failure that is not a closed server as fatal.

### Consumer lifecycle

Build repositories' backing pool and the Redis client. Let each module register handlers. Block in `Run` until the context is cancelled. The read must be a blocking pop with the context (`BLPOP` or equivalent) so an idle process does not spin. A handler runs after a successful pop. See section 16 for the acknowledgement rule.

The reference loop that polls `LPOP` without waiting is not the pattern to copy. The pattern is a single consumer process, a map of list name to handler, and a blocking read.

### Worker lifecycle

If the feature flag is off, wait on the context and return. If it is on, dial, read, and on disconnect wait with exponential backoff bounded at 30 seconds. Stop immediately when the context is done, including during the wait. Reset the backoff after a connection that starts delivering messages.

Workers call use cases with a background context derived from the process context, not with an HTTP request context.

### CLI jobs

`populate` and similar commands are not daemons. They open the pool, run to completion, and cancel their context. They share seed or use-case code with the server rather than a second implementation.

### Schedulers

There is no cron framework. Do not add an in-process ticker for a rare cleanup unless the operation is truly periodic and owned by a worker package. Prefer a CLI command invoked by the operator or by an external scheduler when the work is operational.

---



## 16. Messaging / Event-Driven Architecture



### Role

Queues decouple two things:

1. **Live UI fan-out.** A write publishes a small JSON fact. A consumer in the HTTP process forwards it to the WebSocket hub.
2. **Follow-up work in another module.** A state change publishes a fact. The other module's consumer calls its own use case (for example creating a dependent record).

Queues are not the system of record.

### Producers

The use case marshals a payload struct from `<module>/entity` and calls `queue.Push`. The router supplies a `Queue` bound to one configured list. A use case that publishes two facts receives two `Queue` values, not one queue with an embedded type switch.

Publish after the local decision is valid. If the database write and the publish must agree, write the database first, then publish. The consumer handles the case where the process dies in between by being safe to run once the row exists, and the missing message is a logged operational gap — do not pretend there is an outbox unless you add one.

### Consumers

`delivery/consumer` registers `<config queue name> → function(ctx, queue.Message)`. The function unmarshals `msg.Payload`. On failure it logs and returns. On success it either calls a use case or sends on a channel the WebSocket handler selects on, with `ctx.Done()` so shutdown does not block forever.

Package-level channels are acceptable for this in-process handoff. They are not a general event bus. Do not use them between modules that run in different goroutines without a defined owner. The HTTP handler that owns the socket is the receiver.

### Message structure

JSON, fields camelCase, struct tags on the payload type. Include the ids the consumer needs so it does not guess. Do not include secrets or full entity graphs.

### Naming

List names are configured, not hard-coded in the producer. The Go constant for a list lives in the Redis section of config. Payload types are `<Entity><Fact>`.

### Retry, idempotency, dead letter

The architecture does **not** include a dead-letter list, a retry counter, or an outbox.

Rules that keep that safe:

- Handlers must be idempotent. Before inserting a follow-up row, check the unique foreign key (`transaction_id` or equivalent). A second delivery must no-op or update, not duplicate.
- Unique constraints are the backstop.
- A handler that does required work should be written so the message is acknowledged only after the handler returns nil. If the current loop pops first, new required-work consumers must still tolerate loss and duplication: loss is recovered by an operator or by the state of the source row, duplication is absorbed by the unique key.
- Poison messages (bad JSON) are logged and discarded. They must not block the list.
- Do not retry validation failures.



### Transaction and event boundary

```text
use case validates
use case writes its own row
use case publishes a fact that other modules need
consumer of the other module writes its own row
```

There is no distributed transaction. Cross-module invariants that cannot tolerate a gap need a unique key and a consumer that can be run again, or they need to be the same repository transaction inside one module. Prefer the latter when the two tables are one operation.

---



## 17. Observability



### Logging

- One JSON logger in `pkg/log`, source enabled, three named levels.
- Access logs come from the HTTP framework's default middleware.
- Log at the boundary that first handles an unexpected error (section 11).
- Fields: `error`, plus an id or business key. No credentials, tokens, password, or one-time code.
- The log file path is a process concern. Do not open extra log files per module.



### Metrics

Not part of the current design. If added later, expose them from `infrastructure/server` and increment them in adapters, not in entities.

### Tracing

Not part of the current design. Request ids are not propagated. Do not add a tracer inside a single feature.

### Health

There is no `/health` or `/ready` route today. A new project using this template should add two unauthenticated routes on the HTTP server, implemented in `infrastructure/server` or `httpconfig`, not in a business module:

- **Liveness:** the process is running.
- **Readiness:** the database pool answers a ping. Redis ping belongs here only if sessions and queues are required for serving.

These routes must not use `AuthRequired` and must not touch module use cases.

### Frontend connection status

Browser pages that depend on a WebSocket may show a disconnected state in static JS. That is a presentation detail, not a server health check.

### Audit

There is no audit table. Durable history is the row itself (`created_at`, `updated_at`, status). Add an audit log only as its own table and repository when a requirement names the events to retain. Do not log audit events only to the log file.

---



## 18. Security



### Authentication

- Identity is a server-side session in Redis, referenced by a cookie. The session id secret comes from the environment.
- Login is a use case: load the account, compare the password with `Hash`, return a generic failure message that does not reveal which field was wrong.
- One-time codes are random, stored in Redis with a TTL from config, deleted on success, and rejected when missing or mismatched. The sender is skipped when its `Active` flag is false. Codes are never written to logs.
- Logout clears the session and saves it.



### Authorization

- Route middleware enforces the coarse class: anonymous, any authenticated user, operator-or-higher, end user.
- Role rank is an enum concern (lower rank means more privilege) so handlers do not compare raw strings in new ways. Use the enum helpers.
- The use case enforces the fine rule, including "this actor may update this subject", using the actor loaded from the database.
- Middleware and use case must agree. A route that only hides a button is not authorization.



### Secret management

- Secrets are environment variables. See section 10.
- `.env.example` contains placeholders.
- Session secret, database URL, provider keys, and object-storage keys are not written to yaml, logs, or seed data.
- The production image contains the binary, `config/`, and `frontend/`. It does not contain `.env`. Compose injects the env file at runtime.



### Password storage

bcrypt hashes only. Repositories persist `[]byte`. Request models hold the clear-text password long enough to call the use case. View models never include the hash.

### Encryption in transit

Terminate TLS at the reverse proxy in production. The application listens on HTTP inside the private network. Cookies should be marked secure when the proxy terminates TLS; set that on the session store options when deploying behind HTTPS.

### Input

- JSON binding and binding tags for shape.
- Semantic checks in the use case.
- SQL parameters for every user value.
- HTML templates escape by default. Do not mark user content as safe HTML.
- Uploaded files are read from the multipart header in the handler, mapped to `base.FileSave` (reader, size, content type, extension), and passed to the use case. The use case decides the bucket and object name. Do not use the client's filename as the object key; generate a name and keep the extension.



### Authorization of object storage

Public buckets only for public assets. The storage prefix on the reverse proxy must not become a way to read private buckets.

### Dependencies and admin commands

The privileged-user CLI command reads the password from the terminal without echo, confirms it, and calls the same use case as account creation. It does not insert SQL from the command.

### Operational exposure

The object-storage console and database ports are published in development compose. Production compose should publish only the reverse proxy, unless an operator explicitly needs a port. Do not add the database port to a public host by habit.

---



## 19. Development Workflow



### Local loop

```text
copy .env.example to .env and fill local placeholders
make up                # database, redis, object storage
make migrate           # apply SQL
make populate          # optional fixtures
make run               # Air rebuilds on .go and .html, runs `serve`
```

Air builds `./cmd` and passes `serve`. HTML template changes reload the process. Static files do not need a rebuild.

### Adding a feature

```text
New feature
    ↓
Migration, if state changes
    ↓
Entity, enum, message codes
    ↓
Repository interface and SQL
    ↓
Use case
    ↓
Handler model, handler, route constants, router registration
    ↓
Template and, if it mutates, static/js/request script
    ↓
Consumer, only if another process must react
    ↓
Seed fixtures and tests for the repository and the use case
```



### Checks before handing off

- `make test` for the packages touched.
- A migration has a down section.
- A new path exists only in `httpconfig`.
- A new failure mode is a new code in the module block, asserted by a test.
- `go` vet / the project's usual compile via Air or `go test` is clean.



### Production shape

```text
build a static binary
image copies binary, config/, and frontend/
compose runs proxy + app + postgres + redis + object storage
proxy forwards / to the app and /storage/ to object storage
app command is `<binary> serve`
```

There is no separate frontend deploy.

### Continuous integration

The reference project has no CI workflow. A new project should run, on every change:

1. `go test ./...` with access to a Docker daemon, because integration tests start containers.
2. `gofmt` / vet.
3. A migration apply against a throwaway database is already what tests do; do not invent a second migration check that drifts.

Do not require a live provider, a real Redis, or production credentials in CI.

---



## 20. How To Add A New Feature

Example: a new resource `<feature>` inside an existing module `<module>`. A new business area uses a new directory `internal/apps/<module>` with the same interior.

```text
<module>/
├── entity/<feature>.go
├── enum/<feature>.go                  # only if it has a closed set
├── message/message.go                 # add success codes
├── message/errors.go                  # add failure codes in this module's block
├── params/<feature>param.go
├── repository/<feature>repo.go
├── repository/<feature>repo_test.go
├── usecase/<feature>usecase.go
├── usecase/<feature>usecase_test.go
├── delivery/http/handler/<feature>handler.go
├── delivery/http/handler/model/<feature>model.go
└── delivery/http/router/<feature>router.go

frontend/templates/<module>/<feature>-list.html
frontend/static/js/request/<feature>-request.js
internal/infrastructure/database/migrations/<timestamp>_create_<features>_table.sql
internal/seed/<feature>seed.go
internal/httpconfig/urlpath.go         # constants and resolvers only
```



### Where each part goes

1. **Migration.** Create the table, keys, and enum type. Down drops them in reverse order.
2. **Entity.** Fields, `base.Entity`, and a list struct. No JSON request tags.
3. **Enum.** If a column is a closed set, add the Go enum and the SQL enum in the same change.
4. **Messages.** Success and failure values. Codes use this module's numeric block.
5. **Params.** Use-case param with page and filters. Repository param with limit, offset, and filters.
6. **Repository.** Interface and SQL. `Exist` and `Read` return zero plus nil when nothing matches. List query returns the count.
7. **Use case.** Check the actor, validate, call the repository, map errors to messages, set `UpdatedAt` from `timezone.NowUTC()` on update.
8. **Route constants.** Add `/<module>/<feature>/list/` and the mutation paths. Add a resolver if the path has an id.
9. **Handler model.** Request struct with binding tags. View model that formats display fields.
10. **Handler.** Bind, map, call the use case, render with the composition module or write JSON.
11. **Router.** Construct the repository and use case in the module router. Register routes with `AuthRequired` and the role middleware. Do not construct them in `httpconfig`.
12. **Template.** Extend `base`. Links come from `Routes` or resolvers. The form `action` is the path constant rendered into the page.
13. **Script.** `sendRequest(form, boxID, JSON.stringify({...}), { httpMethod: "POST" })`. The helper reloads on success and shows `message.description` on failure.
14. **Seed and tests.** Add a named fixture. Repository test covers insert, duplicate, and read. Use-case test covers success and each message you added.
15. **Server registration.** If `<module>` is new, call `New<Module>Router(router, panel).Register()` from `infrastructure/server`. If it is an existing module, `Register` already runs; the new resource method is called from it.

A read-only page does not get a repository method that writes. A mutation does not get a template that embeds the decision the use case should return.

---



## 21. How To Add A New External Integration

1. **Name the capability, not the vendor,** unless the vendor's protocol is the whole feature (`sender`, `filemanager`, `<provider>`).
2. **Define the interface** in that package with the methods the use case needs. No vendor types in the method signatures. Accept and return entities, scalars, and `io.Reader`.
3. **Implement it** in the same package. Keep the client struct private. Add `var _ Interface = (*impl)(nil)`.
4. **Configure it** with a section on `Config`: endpoint, credentials from the environment, and an `Active` flag. Document the variables in `.env.example` with placeholders.
5. **Wire it** in the module router or the worker. Pass the interface into the use case. The use case does not call `New<Vendor>`.
6. **Map errors** inside the adapter to a simple `error`. The use case logs it and returns the server message, or a module message if the failure is something the user can act on (rejected address, unknown upstream id).
7. **Suppress or mock.** Provide a mock implementation in the same package for tests. Honor `Active == false` at the adapter edge.
8. **Test** the use case with the mock. Test a pure parser with table tests. Do not call the real vendor from `go test`.
9. **If it is a long-lived connection,** put the loop in `infrastructure/worker`, start it from `serve` under the wait group, and reconnect as in section 15. The worker maps the foreign message and calls one use case method.
10. **If the integration delivers webhooks,** add a route in the owning module, verify the signature in the handler or a tiny adapter, and call the use case with a mapped struct. Do not let the provider's JSON become the entity without an explicit mapping.
11. **Do not** put provider URLs, keys, or payload samples that contain real customer data in the repository.

A second implementation (a different vendor) replaces the struct inside the same package or behind the same constructor switch on config. It does not create a new use case.

---



## 22. Anti-Patterns

These show up easily in this architecture. Do not introduce them.

- **Business rules in handlers.** Role checks that decide a write, price calculations, and "already exists" branches belong in the use case. The handler binds and renders.
- **SQL or driver calls in use cases and handlers.** Even a one-line query. Add the repository method.
- **Infrastructure types in entities.** No `pgx` rows, no Redis commands, no `gin.Context`, no vendor client on an entity.
- **HTTP models as the repository input.** Request structs stay in `handler/model`. Repository params stay in `params`.
- **A shared** `httpconfig` **that constructs repositories and use cases** to enforce one resource's policy. Inject the use case from the module router.
- **Importing another module's** `delivery` **package.** That couples routes and creates cycles. Call the repository or use case from your own router if you must orchestrate.
- **Package import cycles** between `entity` and another module's `entity`. Break them with an id field or by moving the shared enum to a package that does not import either entity.
- `pkg` **importing** `internal`**.** Move that helper under `internal`.
- **Duplicated rules.** A status transition written in the handler and the use case, or a price formula in the template and the use case. One implementation, in the use case or a pure function it calls. Templates may format; they may not decide.
- **God module.** The composition module reads for the shell. It does not become the place new writes go because "it already has the database".
- **Interfaces with no second caller and no test boundary.** Do not wrap a use case in an interface by default. Do interface the repository and the external port.
- **Generic repository, service locator, or global use-case registry.** Constructors are explicit.
- **Stringly SQL.** Concatenating user input, including ids, into query text. Bind parameters. Developer-supplied column names are the only concatenation allowed, and they are literals in Go.
- **Floats for quantities.** Use integers for minor units or `NUMERIC` / decimal.
- **Logging secrets and returning driver errors to the browser.**
- **A second source of route strings** in JavaScript or templates that is not rendered from `httpconfig`.
- **Queue as the only copy of a write.** The row is committed in the owning module. The queue notifies.
- **Popping a required message and ignoring the handler error,** then assuming the follow-up exists. Make the follow-up idempotent, and do not acknowledge before success for new required consumers.
- **Busy-loop polling** of Redis. Block on the list.
- **New process-wide mutable maps** for request data. Package-level channels are only for the WebSocket handoff described in section 16.
- **Tests that reimplement the repository with a fake** and then pass while the SQL is wrong.
- **Copying an inconsistency because it is nearby.** Examples to avoid in new code: a consumer package named `consumers` in one module and `consumer` in another (use `consumer`); calling `GetConfig()` inside a feature; doubling the module name in a path (`/<module>/<module>/...`).

---



## 23. Architecture Rules / Non-Negotiable Rules

These rules are for an AI agent or a developer changing the code.

1. You MUST put a new business operation in `<module>/usecase` and the SQL for it in `<module>/repository`.
2. You MUST return `messages.Message` from use cases for every expected outcome, using a code from that module's block.
3. You MUST keep user-facing text in `<module>/message`.
4. You MUST register HTTP paths as constants in `httpconfig` and use those constants in the router.
5. You MUST apply authentication and role middleware on the route and MUST enforce resource authorization in the use case.
6. You MUST pass repositories, queues, hashers, senders, and file managers into constructors. You MUST NOT new them inside a use-case method.
7. You MUST define a repository or external port as an interface next to its implementation, with a compile-time assertion.
8. You MUST write a Goose migration with `Up` and `Down` before adding a column or table, and you MUST update the repository scan lists to match.
9. You MUST use bound SQL parameters for user values and `NUMERIC` or integers for quantities.
10. You MUST store times in UTC and convert at the presentation boundary.
11. You MUST treat a missing database row as a zero entity and `nil` error inside repositories, and as a not-found message inside use cases.
12. You MUST log unexpected errors without secrets, and MUST respond to the client with the generic server message.
13. You MUST add a repository test and a use-case test for new behavior, using `seed` fixtures and the existing `TestMain` harness when the database is involved.
14. You MUST keep `pkg` free of `internal` imports.
15. You MUST shut down new long-running goroutines when the process context is cancelled.
16. You MUST make a new queue consumer idempotent if it inserts or mutates data.
17. You SHOULD pass `*config.Config` in from the router rather than calling `GetConfig` in feature code.
18. You SHOULD use a blocking pop in the queue loop.
19. You SHOULD keep one module's delivery package from importing another's.
20. You SHOULD put pure, reused rules in `<module>/validation` and unit-test them without Docker.
21. You SHOULD render HTML through the composition module so the shell stays consistent.
22. You SHOULD prefer extending an existing module over creating a new one when the data and the actor are the same.
23. You MUST NOT put business rules only in handlers, templates, or JavaScript.
24. You MUST NOT import a database driver, Redis, or an HTTP framework from `entity` or `enum`.
25. You MUST NOT let `httpconfig` construct repositories.
26. You MUST NOT return raw driver or vendor errors in an HTTP body.
27. You MUST NOT commit `.env`, credentials, tokens, or production URLs that embed secrets.
28. You MUST NOT add a float column for money or measured amounts.
29. You MUST NOT introduce an ORM, a command bus, or a new top-level architecture for a single feature.
30. You MUST NOT create a repository mock that replaces SQL in the standard use-case test.
31. You MUST NOT publish to a queue as a substitute for the write that must be durable.
32. You MUST NOT log passwords, session secrets, API keys, or one-time codes.
33. You SHOULD NOT add a use-case interface with one implementation.
34. You SHOULD NOT add a global variable for request state.
35. You SHOULD NOT duplicate a route string, an error code, or a status transition.

---



## 24. AI Agent Instructions

When you add or change code in a project that follows this template:

1. Read `PROJECT_ARCHITECTURE_TEMPLATE.md` before creating files. Read the module you are changing, including its router, use case, repository, and message package.
2. Follow the directory structure in section 2. Use the names in section 13.
3. Do not invent a new architectural pattern (new layer, new framework, new bus, new ORM) unless the existing seams cannot express the change. If you believe they cannot, explain the conflict and wait for confirmation before implementing it.
4. Preserve dependency direction from section 5. Run a mental import check: entities do not import delivery, `pkg` does not import `internal`, handlers do not import repositories.
5. Put each new type in the layer that already owns that responsibility. A validation rule does not go in the handler because the handler file is the one you have open.
6. Follow naming conventions exactly, including file names without underscores and message codes in the module's numeric block.
7. Add tests for the new behavior at the layer section 12 specifies. Update `seed` when the new behavior needs a row.
8. Do not bypass an existing interface, message type, route catalog, or composition renderer without a stated reason. Wire new dependencies the way the module router already wires the others.
9. If the request conflicts with these rules, stop and explain the conflict before writing the code. Do not silently place SQL in a handler to "match a nearby shortcut".
10. Do not copy project-specific business language, sample secrets, or production hostnames into the new code. Do not print secrets from config into logs while debugging.
11. Keep changes limited to the feature. Do not reformat unrelated packages or rename the existing consumer package as a drive-by.
12. When you finish, state which layer each new file belongs to and which tests cover it.

---



## 25. New Project Bootstrap

1. **Create the repository.** Initialize a Go module. Add a `.gitignore` for `.env`, binaries, `tmp/`, coverage files, and frontend dependency directories. Commit `.env.example` with placeholders only.
2. **Create the directory structure** from section 2, including one real module `<module>` with empty `entity`, `message`, `repository`, `usecase`, and `delivery` packages, plus a composition module for the shell.
3. **Initialize dependencies.** HTTP router, template engine, session store, Postgres driver, Redis client, migration CLI, password hash library, decimal library, CLI library, structured logging, test assertions. Pin them in `go.mod`.
4. **Configure the application.** Add `config.Config` with yaml and environment loading, `config.yaml` for non-secrets, and `.env.example`. `GetConfig` panics on a missing required value. Wire Air to build `./cmd` and run `serve`.
5. **Configure infrastructure.** Compose file for Postgres, Redis, and object storage. `infrastructure/database` pool, `infrastructure/cache` client, `infrastructure/queue` list wrapper, `infrastructure/hash`, and stub `sender` and `filemanager` interfaces even if the first feature uses only the database.
6. **Create domain boundaries.** Add `internal/types` (`ID`, nullables) and `internal/utils/base` (`Entity`, `Pagination`). Add `pkg/errors` and `internal/utils/messages`. Create the first module's entity and enum only for the first aggregate. Choose its error-code block.
7. **Create the application layer.** One use case with `Create` and `Read`, returning messages. Repository interface in the same change, implemented with parameterized SQL.
8. **Create the transport layer.** `httpconfig` route constants, response envelope, template renderer, session middleware, `AuthRequired`. Module router registers routes. `infrastructure/server.Run` builds the engine, static files, templates, and the router bag, then calls `Register`. Add `frontend/templates/base` and one page. Add `frontend/static/js/request` with the shared `sendRequest` helper.
9. **Add the database and migrations.** First Goose file creates the first tables. Embed the directory for tests. `Makefile` targets: `up`, `down`, `migrate`, `migrate-down`, `migration`, `run`, `test`, `populate`.
10. **Add the testing infrastructure.** `pkg/testdocker` for Postgres. `database.SetupTestDB`. `internal/seed` with one fixture. `TestMain` in the repository and use-case packages. One unit test for a pure helper so the pattern is visible without Docker as well.
11. **Add observability and process control.** `pkg/log` JSON logger. HTTP server graceful shutdown. `serve` starts the server under a context cancelled by `SIGINT` and `SIGTERM`. Add unauthenticated liveness and readiness routes in the server package.
12. **Add delivery extras when the first feature needs them.** Queue consumer process and a module `delivery/consumer` only when something must happen asynchronously. A worker only when an external stream exists. Production Dockerfile that copies the binary, `config/`, and `frontend/`. Compose file that runs the reverse proxy, the app, and the data stores. CI that runs `go test ./...` with Docker available.
13. **Run the loop once before adding features.** `make up`, `make migrate`, `make test`, `make run`, open the first page, submit the first JSON mutation, and confirm the envelope and the row.

After bootstrap, every feature follows section 20. Do not add a second service, a second module path, or a public API framework until a requirement cannot be met inside this process.

---



# Quick Reference



### Directory → responsibility


| Path                                       | Responsibility                                    |
| ------------------------------------------ | ------------------------------------------------- |
| `cmd/`                                     | CLI and process start                             |
| `config/`                                  | Typed config, yaml + env                          |
| `frontend/templates/`                      | HTML pages and layout                             |
| `frontend/static/`                         | CSS, JS, images                                   |
| `internal/apps/<module>/entity`            | Records and queue payloads                        |
| `internal/apps/<module>/enum`              | Closed string sets                                |
| `internal/apps/<module>/message`           | Success and failure results                       |
| `internal/apps/<module>/params`            | Use-case and repository inputs                    |
| `internal/apps/<module>/repository`        | SQL and cache access                              |
| `internal/apps/<module>/usecase`           | Application operations                            |
| `internal/apps/<module>/delivery/http`     | Handlers, view models, routes                     |
| `internal/apps/<module>/delivery/consumer` | Queue handlers                                    |
| `internal/apps/<composition-module>/`      | Shell read model and HTML render                  |
| `internal/httpconfig/`                     | Path catalog, envelope, auth middleware, renderer |
| `internal/infrastructure/database/`        | Pool and migrations                               |
| `internal/infrastructure/queue/`           | Redis lists and the consumer loop                 |
| `internal/infrastructure/server/`          | HTTP and consumer processes                       |
| `internal/infrastructure/worker/`          | External long-lived clients                       |
| `internal/seed/`                           | Fixtures                                          |
| `internal/types/`                          | IDs and nullables                                 |
| `internal/utils/`                          | Entity base, pagination, message type             |
| `pkg/`                                     | Errors, log, time, validators, test containers    |




### Layer → responsibility


| Layer          | Responsibility                          |
| -------------- | --------------------------------------- |
| Process        | Wire and run                            |
| Transport      | HTTP, HTML, JSON, sessions              |
| Application    | Authorize, decide, orchestrate          |
| Module model   | Values, enums, pure rules, result codes |
| Persistence    | Queries                                 |
| Infrastructure | Postgres, Redis, vendors                |
| Composition    | Shared page shell                       |




### Dependency direction

```text
cmd → server → router → handler → usecase → repository interface
                                 → port interface (queue, hash, sender, files)
repository implementation → driver
entity → types, base
pkg ↛ internal
delivery ↛ another module's delivery
```



### Where new features go

`entity` → `message` → `repository` → `usecase` → `handler/model` + `handler` → `httpconfig` path → module `router` → template and `static/js/request` → migration and `seed`.

### Where tests go

Next to the package, external test package. Use cases and repositories: Postgres via `SetupTestDB` and `seed`. Pure rules: table-driven unit tests. Not in a separate `tests/` tree.

### Where integrations go

Interface and implementation in `internal/infrastructure/<concern>`. Constructed in the module router or a worker. Use case depends on the interface.

### Where database code goes

Schema: `internal/infrastructure/database/migrations`. Queries: the owning module's `repository`. Shell-only reads: the composition repository. Pool: `infrastructure/database`.

### Where API code goes

Paths: `internal/httpconfig`. Registration and middleware: `<module>/delivery/http/router`. Behavior: `<module>/delivery/http/handler`. JSON and view models: `handler/model`. Pages: `frontend/templates/<module>`. Mutations from the browser: `frontend/static/js/request`.

### Critical architecture rules

- Handlers do not own rules or SQL.
- Use cases return messages; they do not return driver errors to HTTP.
- Repository interfaces live with their SQL; external ports live with their adapters.
- One route catalog. One error-code block per module.
- Database is the source of truth; queues notify or continue work and must be idempotent.
- `pkg` does not import `internal`.
- No floats for quantities, no secrets in the repo, no new framework for one feature.

