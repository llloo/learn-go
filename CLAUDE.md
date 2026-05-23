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

## Current project state (Phase 2 complete)

```
taskapi/
├── cmd/server/main.go          # Entry point (chi router)
├── go.mod
├── go.sum
├── handler_test.go
├── internal/
│   ├── task/task.go            # Task struct
│   ├── store/store.go          # TaskStore interface + in-memory Store
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
cmd/server/main.go  →  handler.Server  →  store.TaskStore (interface)
     │                    (depends on)       ↑ implements
     │r.Use(Logger)                    store.Store (in-memory)
     │
     └── Logger (middleware) → logs Method + Path
```

- `handler.Server.Store` uses interface `store.TaskStore` — swap implementations without changing handler
- chi router handles method routing (`r.Get`, `r.Post`) and path params (`chi.URLParam(r, "id")`)
- Middleware pattern: `func(next http.Handler) http.Handler`, registered via `r.Use()`
- Error handling: `APIError` struct + `WriteError` helper → consistent JSON `{"message": "..."}` responses
- Packages layered: `cmd/server` → `internal/handler` → `internal/store` → `internal/task`
