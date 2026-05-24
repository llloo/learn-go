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
go vet ./...                 # Vet
go fmt ./...                 # Format
```

## Current project state (Phase 3 complete)

```
taskapi/
├── cmd/server/main.go          # Entry point: config → migrate → PostgresStore → chi
├── go.mod
├── go.sum
├── handler_test.go             # Handler tests (uses in-memory Store as fake)
├── migrations/
│   ├── 000001_create_tasks.up.sql
│   └── 000001_create_tasks.down.sql
├── internal/
│   ├── task/task.go            # Task struct
│   ├── config/config.go        # envconfig: SERVER_PORT, DATABASE_URL
│   ├── store/
│   │   ├── store.go            # TaskStore interface + in-memory Store (test fake)
│   │   └── postgres.go         # PostgresStore: pgx + database/sql
│   └── handler/
│       ├── handler.go          # HTTP handlers (chi)
│       ├── middleware.go       # Logger middleware
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
    ├── config.NewConfig()          → envconfig binds env vars
    ├── migration()                 → golang-migrate runs *.sql files
    ├── store.NewPostgresStore()    → PostgresStore (implements TaskStore)
    ├── handler.Logger              → middleware: func(next http.Handler) http.Handler
    ├── handler.Server              → depends on store.TaskStore (interface)
    └── chi router                  → r.Get/Post + chi.URLParam
```

- `handler.Server.Store` uses interface `store.TaskStore` — swapped from memory to PostgresStore without handler changes
- PostgresStore uses `pgx` driver via blank import `_ "github.com/jackc/pgx/v5/stdlib"`
- `context.Context` flows from `r.Context()` through every store method
- `errors.Is(err, sql.ErrNoRows)` to distinguish "not found" from real errors
- Migrations run on startup via `golang-migrate`, idempotent (`ErrNoChange` handled)
- Error handling: `APIError` struct + `WriteError` helper → consistent JSON `{"message": "..."}` responses
- Packages layered: `cmd/server` → `internal/handler` → `internal/store` → `internal/task`
- `internal/config` is independent, used only by `main.go`
