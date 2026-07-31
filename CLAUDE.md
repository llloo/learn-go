# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Go learning repository for a senior Python engineer. The learning path uses a **single growing project** (Task REST API) structured in 5 phases plus a JWT auth extension. Full spec: [docs/superpowers/specs/2026-05-19-go-learning-path-design.md](docs/superpowers/specs/2026-05-19-go-learning-path-design.md)

## User context

- Deep Python expertise, new to Go
- Frame Go concepts against Python analogues
- Project-driven learning style — build first, explain concepts as they arise in code
- Goal: backend API / microservices with Go

## Design principles for this project

- Standard library first, third-party packages only when the stdlib approach is understood
- Interface-driven design from Phase 2 onward
- Table-driven tests (Go convention, differs from pytest parametrize)
- No web frameworks (Gin, Echo) — learn `net/http` and `chi` fundamentals first

## CRITICAL: Do NOT write code for the user

This is a project-driven **learning** repository. Explain concepts, show examples, guide the approach — but do **not** write or edit Go files directly unless the user explicitly asks you to. If uncertain, ask first.

## Go environment

Go is installed at `/usr/local/go/bin/go` (1.26.3, already in PATH via `.bashrc`).

## Common Go commands

```bash
go build ./...               # Build all packages
go test ./...                # Run all tests
go test -run TestName ./...  # Run a single test
go test -bench=. -benchmem ./...  # Run benchmarks
go vet ./...                 # Vet
go fmt ./...                 # Format
make build                   # Build binary via Makefile
make test                    # Run tests via Makefile
make lint                    # golangci-lint
```

## Current project state (5 phases + JWT auth complete)

The repository root **is** the Go module (`module taskapi` in `go.mod`) — files live at the root, not in a `taskapi/` subdirectory.

```
.  (module: taskapi)
├── cmd/server/main.go          # Entry point: config → migrate → PostgresStore → auth → chi
├── Dockerfile                  # Multi-stage (golang:1.26-alpine → alpine)
├── .dockerignore
├── Makefile                    # build, test, bench, vet, lint, docker-build, clean
├── .golangci.yml               # errcheck + govet + ineffassign
├── go.mod / go.sum
├── handler_test.go             # Black-box tests, `package main` at repo root (imports internal/*)
├── handle_bench_test.go        # Benchmark tests (serial vs concurrent)
├── migrations/
│   ├── 000001_create_tasks.up.sql
│   └── 000001_create_tasks.down.sql
├── internal/
│   ├── task/task.go            # Task struct
│   ├── config/config.go        # envconfig: APP_SERVER_PORT, APP_DATABASE_URL, APP_JWT_SECRET
│   ├── auth/auth.go            # auth.Service: JWT HS256 generate/validate (golang-jwt/v5)
│   ├── store/
│   │   ├── store.go            # TaskStore interface + in-memory Store (test fake)
│   │   └── postgres.go         # PostgresStore: pgx + database/sql
│   └── handler/
│       ├── handler.go          # CRUD handlers + Server struct (Store + Auth)
│       ├── auth.go             # HandleUserLogin
│       ├── batch.go            # BatchResult + HandleBatchCreateTasks
│       ├── middleware.go       # Logger + RequireAuth (JWT) middleware
│       └── error.go            # APIError + WriteError
└── docs/
    ├── python-go-cheatsheet.md
    ├── superpowers/specs/2026-05-19-go-learning-path-design.md
    └── superpowers/plans/2026-05-20-phase-1-fundamentals.md
```

## Architecture

```
cmd/server/main.go
    │
    ├── config.NewConfig()              → envconfig binds env vars (incl. APP_JWT_SECRET)
    ├── slog.NewJSONHandler()           → structured JSON logging
    ├── migration()                     → golang-migrate runs *.sql files
    ├── store.NewPostgresStore()        → PostgresStore (implements TaskStore)
    ├── auth.NewService()               → JWT sign/verify (HS256)
    ├── handler.Server                  → {Store: TaskStore, Auth: *auth.Service}
    ├── handler.Logger (middleware)     → slog: method + path
    ├── handler.RequireAuth (middleware)→ validates Bearer token, injects user_id into ctx
    ├── chi router                      → r.Group protected vs public routes
    └── signal.NotifyContext            → graceful shutdown on SIGINT/SIGTERM

Routes:
    PUBLIC     GET /tasks, GET /tasks/{id}, POST /auth/login
    PROTECTED  POST /tasks, POST /tasks/batch, PATCH /tasks/{id}, DELETE /tasks/{id} (RequireAuth)
```

### Concurrency (Phase 4)
- `POST /tasks/batch` — goroutine per title + buffered channel + `select` timeout
- Semaphore pattern: `make(chan struct{}, N)` limits concurrent goroutines
- `select`: channel result / context.Done() / time.After
- Benchmark: serial faster for in-memory (mutex contention), concurrent wins for I/O

### Infrastructure (Phase 5)
- `log/slog` structured JSON logging throughout main.go and middleware
- `slog.Error` + `os.Exit(1)` manual exit (slog doesn't exit like log.Fatal)
- `signal.NotifyContext`: context cancelled on SIGINT/SIGTERM → `srv.Shutdown()`
- Server runs in goroutine, main blocks on `<-ctx.Done()`
- Docker multi-stage build: `golang:1.26-alpine` (builder) → `alpine` (runtime, ~15MB binary)
- Makefile: `make build/test/bench/vet/lint/docker-build/clean`
- `golangci-lint` with `errcheck`, `govet`, `ineffassign` linters

### Authentication (Phase 6, JWT)
- `auth.Service` (internal/auth): `GenerateToken(userID)` / `ValidateToken(tokenStr)`
  - `golang-jwt/jwt/v5`, HS256, `jwt.MapClaims` with `user_id` + `exp` (24h)
  - ValidateToken verifies signing method is HMAC (alg-confusion guard) then returns `user_id`
- `handler.RequireAuth` middleware: `Authorization: Bearer <token>` → `ValidateToken` → `context.WithValue(ctx, UserIDKey, userID)` → `handler.GetUserID(ctx)`
- `contextKey` typed string type — idiomatic Go way to avoid context-key collisions (Python analog: enum/keys module)
- `POST /auth/login` returns a token; **stub**: hardcoded `user123`, no password check, no user table
- `Server` struct now holds `{Store, Auth}` — middleware and handlers share the same dependency injection point
- JWTSecret via `APP_JWT_SECRET` env var (has a `"your-secret-key"` default — fine for learning, change for prod)

### Known gaps (learning opportunities, not bugs)
- JWT auth has **no tests yet** — `internal/auth` and `RequireAuth` are untested
- Login is a stub (no real users/passwords); `exp` is not validated explicitly by RequireAuth (jwt.Parse does it via claims validation)
- Root-level tests are `package main` in a directory with no non-test .go files — builds fine, but unconventional; a normal black-box layout would put them in `internal/handler` as `package handler_test`

### Core patterns (all phases)
- `handler.Server.Store` uses interface `store.TaskStore` — swap memory/PostgresStore
- `context.Context` flows from `r.Context()` through every store method
- `errors.Is(err, sql.ErrNoRows)` to distinguish "not found" from real errors
- Migrations run on startup via `golang-migrate`, idempotent (`ErrNoChange` handled)
- Error handling: `APIError` struct + `WriteError` helper → consistent JSON `{"message": "..."}` responses
- Packages layered: `cmd/server` → `internal/handler` → `internal/store` → `internal/task`; `internal/auth` sits beside handler (used by middleware + login handler)
- `internal/config` is independent, used only by `main.go`
- Context values: typed `contextKey` + `context.WithValue` + `GetUserID(ctx)` accessor — values cross handler→middleware boundaries without globals
