# Phase 7 Design: Real Authentication System

Date: 2026-08-01
Status: Approved (brainstormed with user)

## Context

The Task REST API has completed all 5 original learning phases plus a Phase 6 JWT auth extension. Phase 6 left deliberate gaps, recorded in CLAUDE.md as learning opportunities:

- `internal/auth` and `RequireAuth` middleware have **no tests**
- `POST /auth/login` is a stub: hardcoded `user123`, no password check, no user table
- No real user persistence or credential verification

This phase replaces the stub with a real authentication system while applying the project's established patterns for the second time (`TaskStore` interface pattern → `UserStore`).

## Learning Goals (this phase)

- Database schema evolution: a second migration (`000002`) on top of the existing one
- Password hashing with bcrypt (`golang.org/x/crypto/bcrypt`): salt, cost factor, constant-time comparison
- Database error code → domain error mapping (`23505` unique violation → `store.ErrConflict`)
- Account-enumeration prevention: identical 401 responses for "unknown user" and "wrong password"
- Second application of the interface + fake + Postgres implementation pattern (`UserStore`)
- Table-driven tests for auth service, middleware, and handlers

## Roadmap Overview (Phase 7–10)

Phases are executed sequentially; each is spec'd and planned independently.

| Phase | Topic | Notes |
|---|---|---|
| 7 (this) | Real authentication | Users table, bcrypt, register/login, auth tests |
| 8 | Production polish | Rate limiting (`golang.org/x/time/rate`), pprof, docker-compose, GitHub Actions CI |
| 9 | Go language depth | errgroup, context timeouts, generics — **woven into phases 7–8**, not a standalone phase |
| 10 | Toward microservices | Second service (HTTP client with timeout/retry, or gRPC); direction re-brainstormed after Phase 8 |

## Phase 7 Design

### Package structure (unchanged layout, additions in bold)

```
internal/
├── auth/auth.go            # JWT (existing) + HashPassword / VerifyPassword (new)
├── user/user.go            # NEW: User struct
├── store/store.go          # TaskStore (existing) + UserStore interface + memory fake (new)
├── store/postgres.go       # PostgresStore (existing) + PostgresUserStore (new)
└── handler/
    ├── auth.go             # HandleUserLogin (rewritten) + HandleRegister (new)
    └── middleware.go       # RequireAuth (unchanged)
```

`auth.Service` owns both JWT and password hashing — it is the "credentials domain" package. No separate `password` package (YAGNI). `UserStore` mirrors `TaskStore`'s structure exactly (interface + fake + Postgres in `internal/store`), so the pattern is learned by repetition.

### Migration `000002_create_users`

```sql
-- up
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- down
DROP TABLE users;
```

- `username UNIQUE` is the data-layer guard for concurrent duplicate registration
- First table with a real constraint (tasks has none) — introduces `23505` error mapping

### Data layer

`internal/user/user.go`:

```go
type User struct {
    ID           int
    Username     string
    PasswordHash string
    CreatedAt    time.Time
}
```

`UserStore` interface (in `internal/store/store.go`, next to `TaskStore`):

```go
type UserStore interface {
    Create(ctx context.Context, username, passwordHash string) (*user.User, error)
    GetByUsername(ctx context.Context, username string) (*user.User, error)
}
```

- Memory fake: `map[string]*user.User` keyed by username + `sync.RWMutex` — mirrors the TaskStore fake pattern
- `PostgresUserStore` in `postgres.go`: `INSERT ... RETURNING`, `WHERE username = $1`, `sql.ErrNoRows` → `store.ErrNotFound`
- New domain error: `store.ErrConflict` — returned by both fake (pre-check map) and Postgres (23505 → ErrConflict). Handler maps it to 409

### auth.Service extensions

```go
func (s *Service) HashPassword(password string) (string, error)
func (s *Service) VerifyPassword(hash, password string) error
```

- bcrypt, default cost factor, `GenerateFromPassword` / `CompareHashAndPassword`
- Python analogue: `bcrypt`/`passlib` — same concept (store hash, constant-time compare)
- JWT methods unchanged; `GenerateToken` now receives the real username

### Endpoints

Both in the public route group:

```
POST /auth/register   (new)
  body: {"username": "...", "password": "..."}
  201 on success, empty body — no token, client must call login afterwards
  400 invalid input (username empty or > 50 chars, password < 8 chars)
  409 username already taken (store.ErrConflict)

POST /auth/login      (rewritten — replaces hardcoded user123)
  body: {"username": "...", "password": "..."}
  200 {"token": "..."} on success (user_id = real username)
  401 for BOTH unknown user and wrong password — identical message (anti-enumeration)
```

Input validation lives in a small `validateCredentials(username, password string) error` helper in the handler package.

### Error handling (learning focus)

1. Domain errors: `store.ErrNotFound` (existing), `store.ErrConflict` (new)
2. Postgres `23505` → `ErrConflict`; `sql.ErrNoRows` → `ErrNotFound` (existing pattern)
3. Handler dispatches with `errors.Is`: 400 (validation) / 401 (auth failure) / 409 (conflict) / 500 (other)
4. Login never distinguishes "user doesn't exist" from "wrong password" at the HTTP layer

### Test plan

All new tests are table-driven (project convention). New files:

- `internal/auth/auth_test.go`
  - Generate/Validate round-trip; expired token (`exp` in the past); tampered token; wrong signing algorithm (RS256/`alg=none` → rejected)
  - Hash/Verify: correct round-trip; wrong password fails
- `internal/handler/auth_test.go`
  - Register: success / duplicate username (409) / short password (400)
  - Login: success / unknown user (401) / wrong password (401)
  - RequireAuth: missing header (401) / wrong format (401) / invalid token (401) / valid token → `GetUserID(ctx)` returns user_id
- `handler_test.go`: `newTestServer` gains an `Auth: auth.NewService(...)` field (currently nil)

Test style: in-memory fakes for stores + real `auth.Service` — no mock library, consistent with existing tests.

### Wiring

`cmd/server/main.go`: register `/auth/register` route next to `/auth/login`. Server struct wiring (`{Store, Auth}`) already exists.

New dependency: `golang.org/x/crypto/bcrypt` (one module).

## Success Criteria

- `POST /auth/login` with hardcoded credentials fails; only registered users can log in
- `POST /auth/register` → login → protected endpoint round-trip works
- Duplicate registration returns 409; wrong credentials return 401
- All existing tests pass; new auth tests pass
- `go vet ./...` and `make lint` clean

## Out of Scope (deferred)

- UUID user IDs (revisit at microservices phase)
- Argon2id / password policy tuning
- Token refresh / revocation / logout
- Seed user in migration (register endpoint makes the API self-contained)
- Rate limiting on login (Phase 8)
