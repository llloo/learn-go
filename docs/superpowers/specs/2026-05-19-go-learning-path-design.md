# Go Learning Path: Project-Driven from Python

## Context

Senior Python engineer learning Go from scratch. Preference for project-driven learning with a concrete project rather than textbook-style progression.

**Learning style:** Build first, understand depth through doing.
**Goal:** Backend API / microservices development with Go.

## Project: Task API

A CRUD REST API for task management — domain is intentionally simple (zero business-learning overhead) so all focus is on Go itself. Every Python back-end engineer has built this before, which means the mental model exists and the difference is purely _how Go does it_.

### Why this project

- CRUD covers ~90% of backend scenarios
- Naturally introduces struct, interface, error handling, concurrency, testing
- Easy to extend without breaking earlier learning
- Direct Python comparison possible at every step

## Phase 1: Fundamentals + Hello HTTP

**Theme:** Get a working HTTP server with in-memory storage using only the standard library. Learn Go syntax by writing real code, not by reading docs.

**Concepts covered:**

- `go mod init`, minimal Go project structure
- Types: `struct`, `slice`, `map`, pointer semantics
- `var` vs `:=`, zero values, exported vs unexported
- `net/http` server and handler registration
- `encoding/json` marshaling with struct tags
- `sync.Mutex` for thread-safe in-memory storage

**Endpoints:** `GET /tasks`, `POST /tasks`, `GET /tasks/{id}`

**Deliverable:** Running HTTP service + Python↔Go cheat sheet

## Phase 2: Project Structure & Routing

**Theme:** Refactor the single-file prototype into a maintainable multi-package layout. Introduce interfaces and middleware — the two Go patterns that differ most from Python.

**Concepts covered:**

- Go project layout: `cmd/`, `internal/`, handler/service/repository layering
- Package philosophy: organize by responsibility, not by type
- `chi` or `net/http.ServeMux` for routing
- `TaskStore` interface for dependency inversion
- Middleware pattern: `func(http.Handler) http.Handler`
- Error handling: `if err != nil`, custom error types, error wrapping with `fmt.Errorf`

**Deliverable:** Clean project skeleton ready for real storage and testing

## Phase 3: Database & Configuration

**Theme:** Replace in-memory storage with PostgreSQL. Introduce real configuration management and the `context.Context` pattern.

**Concepts covered:**

- Environment-based config with `envconfig`
- `database/sql` + `pgx` driver, connection pooling, context timeouts
- Repository pattern — swap `TaskStore` implementation via interface
- Schema migrations with `golang-migrate`
- `context.Context` propagation through call chain

**Deliverable:** Persistent API with clean config and migration tooling

## Phase 4: Concurrency & Testing

**Theme:** Add testing coverage and leverage Go's concurrency primitives. This is where Go's differentiation from Python is most visible.

**Concepts covered:**

- `go test`, table-driven tests (idiomatic Go), `httptest`
- Goroutines vs Python `asyncio` / `threading`
- Channels: buffered/unbuffered, `select`, fan-out/fan-in
- `go test -bench`, pprof basics
- Batch processing endpoint using goroutine pool

**Deliverable:** Tested API with concurrent batch processing capability

## Phase 5: Production Readiness

**Theme:** Everything needed to ship. Logging, graceful shutdown, Docker, CI-ready Makefile.

**Concepts covered:**

- `log/slog` structured logging (Go 1.21+)
- `signal.NotifyContext` graceful shutdown
- Multi-stage Docker build, static binary advantages
- `go vet`, `golangci-lint` static analysis
- Makefile: build, test, lint, migrate targets

**Deliverable:** Production-ready service with Dockerfile and Makefile

## Key Design Decisions

1. **Standard library first, third-party later.** Phase 1 uses `net/http` directly. Phase 2 introduces `chi`. This ensures understanding of what the router abstracts.
2. **Interface-driven from Phase 2.** `TaskStore` is the backbone — memory implementation in Phase 1/2, SQL in Phase 3, mock in Phase 4 tests.
3. **No framework until it hurts.** This path deliberately avoids Gin/Echo/Go-Micro to build fundamentals first.
4. **Table-driven tests as the default.** This is Go community convention and differs from Python's pytest parametrize style.

## Out of Scope (deliberately excluded)

- gRPC / protobuf (separate learning path, not needed for first Go project)
- Generic types (not needed for this project scope)
- Advanced concurrency patterns (worker pools, rate limiting — Phase 4 covers enough to start)
