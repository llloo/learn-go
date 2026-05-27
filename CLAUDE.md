# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Go learning repository for a senior Python engineer. The learning path uses a **single growing project** (Task REST API) structured in 5 phases. Full spec: [docs/superpowers/specs/2026-05-19-go-learning-path-design.md](docs/superpowers/specs/2026-05-19-go-learning-path-design.md)

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

## Current project state (All 5 phases complete)

```
taskapi/
├── cmd/server/main.go          # Entry point: config → migrate → PostgresStore → chi
├── Dockerfile                  # Multi-stage (golang:1.26-alpine → alpine)
├── .dockerignore
├── Makefile                    # build, test, bench, vet, lint, docker-build, clean
├── .golangci.yml               # errcheck + govet + ineffassign
├── go.mod / go.sum
├── handler_test.go             # Handler tests (in-memory Store as fake)
├── handle_bench_test.go        # Benchmark tests (serial vs concurrent)
├── migrations/
│   ├── 000001_create_tasks.up.sql
│   └── 000001_create_tasks.down.sql
├── internal/
│   ├── task/task.go            # Task struct
│   ├── config/config.go        # envconfig: APP_SERVER_PORT, APP_DATABASE_URL
│   ├── store/
│   │   ├── store.go            # TaskStore interface + in-memory Store (test fake)
│   │   └── postgres.go         # PostgresStore: pgx + database/sql
│   └── handler/
│       ├── handler.go          # HTTP handlers (chi)
│       ├── batch.go            # BatchResult + HandleBatchCreateTasks
│       ├── middleware.go       # Logger middleware (slog)
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
    ├── config.NewConfig()              → envconfig binds env vars
    ├── slog.NewJSONHandler()           → structured JSON logging
    ├── migration()                     → golang-migrate runs *.sql files
    ├── store.NewPostgresStore()        → PostgresStore (implements TaskStore)
    ├── handler.Logger (middleware)     → slog: method + path
    ├── handler.Server                  → depends on store.TaskStore (interface)
    ├── chi router                      → r.Get/Post + chi.URLParam + r.Post batch
    └── signal.NotifyContext            → graceful shutdown on SIGINT/SIGTERM
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

### Core patterns (all phases)
- `handler.Server.Store` uses interface `store.TaskStore` — swap memory/PostgresStore
- `context.Context` flows from `r.Context()` through every store method
- `errors.Is(err, sql.ErrNoRows)` to distinguish "not found" from real errors
- Migrations run on startup via `golang-migrate`, idempotent (`ErrNoChange` handled)
- Error handling: `APIError` struct + `WriteError` helper → consistent JSON `{"message": "..."}` responses
- Packages layered: `cmd/server` → `internal/handler` → `internal/store` → `internal/task`
- `internal/config` is independent, used only by `main.go`
