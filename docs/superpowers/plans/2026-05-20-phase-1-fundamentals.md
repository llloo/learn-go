# Phase 1: Fundamentals + Hello HTTP — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A working HTTP server (in-memory storage, 3 endpoints) built with only the Go standard library, teaching Go syntax through real code.

**Architecture:** Single-package flat layout. `main.go` wires the server, `task.go` defines the data model, `store.go` handles thread-safe in-memory storage with a mutex, `handler.go` contains HTTP handlers with manual URL-path parsing.

**Tech Stack:** Go standard library only — `net/http`, `encoding/json`, `sync`, `time`, `strconv`.

---

### Task 1: Install Go

**Files:**
- None

- [ ] **Step 1: Install Go 1.22+**

```bash
sudo apt-get update && sudo apt-get install -y golang-go
```

Verify:
```bash
go version
```
Expected: `go version go1.22.x linux/amd64` or newer.

---

### Task 2: Initialize the Go module

**Files:**
- Create: `go.mod`

- [ ] **Step 1: Init the module**

```bash
cd /home/li/go-project && go mod init taskapi
```

- [ ] **Step 2: Verify**

```bash
cat go.mod
```
Expected: module declaration with `taskapi` and Go version.

---

### Task 3: Define the Task struct

**Files:**
- Create: `task.go`

- [ ] **Step 1: Write task.go**

```go
package main

import "time"

// Task represents a single task item.
// Struct tags tell encoding/json how to marshal field names (camelCase in JSON).
// Compare: Python dataclass with field(alias=...) or Pydantic's Field(alias=...).
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}
```

**Python analogue:**
```python
@dataclass
class Task:
    id: int
    title: str
    completed: bool = False
    created_at: datetime = field(default_factory=datetime.now)
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/li/go-project && go build ./...
```
Expected: builds without errors (no `main` function yet, but the package compiles).

---

### Task 4: Implement thread-safe in-memory store

**Files:**
- Create: `store.go`

- [ ] **Step 1: Write store.go**

```go
package main

import (
	"sync"
	"time"
)

// TaskStore holds tasks in a map protected by a mutex.
// Compare: Python's threading.Lock around a dict.
// The mutex guarantees safe concurrent access — goroutines are green threads,
// and without the lock, concurrent reads/writes to the map would cause a panic.
type TaskStore struct {
	mu     sync.Mutex
	tasks  map[int]Task
	nextID int
}

// NewTaskStore creates an initialized store.
// Go has no constructor — NewXxx is the convention.
func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

// GetAll returns all tasks as a slice, newest first.
// The lock is held only while copying — no lock leaks to callers.
func (s *TaskStore) GetAll() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]Task, 0, len(s.tasks))
	for id := s.nextID - 1; id >= 1; id-- {
		if t, ok := s.tasks[id]; ok {
			result = append(result, t)
		}
	}
	return result
}

// GetByID returns a task and a boolean "ok" (Go's alternative to
// Python's "return None if not found" — zero value + ok is idiomatic Go).
func (s *TaskStore) GetByID(id int) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	return t, ok
}

// Create adds a new task and returns it with the assigned ID and timestamp.
// The caller provides only the title; CreatedAt is set automatically.
func (s *TaskStore) Create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Task{
		ID:        s.nextID,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	s.tasks[s.nextID] = t
	s.nextID++
	return t
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/li/go-project && go build ./...
```
Expected: compiles cleanly.

---

### Task 5: Create HTTP handlers

**Files:**
- Create: `handler.go`

- [ ] **Step 1: Write handler.go**

```go
package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Server bundles dependencies for handlers.
// Compare: FastAPI's dependency injection or Flask's app.config.
type Server struct {
	store *TaskStore
}

// ---- routing glue ----

// tasksHandler handles /tasks (collection: GET list, POST create).
func (s *Server) tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTasks(w, r)
	case http.MethodPost:
		s.createTask(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// taskHandler handles /tasks/{id} (single item).
// Go 1.22's ServeMux supports patterns like /tasks/{id}, but here we
// parse manually to understand URL handling at the lowest level.
func (s *Server) taskHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /tasks/42 → strip prefix "/tasks/" → "42"
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if idStr == "" || strings.Contains(idStr, "/") {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getTask(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ---- handler implementations ----

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.store.GetAll()
	writeJSON(w, http.StatusOK, tasks)
}

// createTaskRequest matches the expected POST body.
type createTaskRequest struct {
	Title string `json:"title"`
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	task := s.store.Create(req.Title)
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request, id int) {
	task, ok := s.store.GetByID(id)
	if !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// ---- helpers ----

// writeJSON sets Content-Type and encodes v as JSON.
// Compare: FastAPI's JSONResponse or Flask's jsonify.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/li/go-project && go build ./...
```
Expected: compiles cleanly.

---

### Task 6: Wire up main.go

**Files:**
- Create: `main.go`

- [ ] **Step 1: Write main.go**

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewTaskStore()
	server := &Server{store: store}

	// Register handlers on the default ServeMux (global router).
	// /tasks matches exactly /tasks; /tasks/ matches everything under it.
	http.HandleFunc("/tasks", server.tasksHandler)
	http.HandleFunc("/tasks/", server.taskHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

- [ ] **Step 2: Build and run**

```bash
cd /home/li/go-project && go build -o taskapi .
```
Expected: binary `taskapi` created.

---

### Task 7: Manual smoke test with curl

**Files:**
- None

- [ ] **Step 1: Start the server in background**

```bash
cd /home/li/go-project && ./taskapi &
sleep 1
```

- [ ] **Step 2: Create a task**

```bash
curl -s -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "learn Go"}'
```
Expected: `{"id":1,"title":"learn Go","completed":false,"created_at":"..."}`

- [ ] **Step 3: Create another task**

```bash
curl -s -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "build something"}'
```

- [ ] **Step 4: List all tasks**

```bash
curl -s http://localhost:8080/tasks
```
Expected: JSON array with 2 tasks, newest first.

- [ ] **Step 5: Get a single task**

```bash
curl -s http://localhost:8080/tasks/1
```
Expected: `{"id":1,"title":"learn Go","completed":false,"created_at":"..."}`

- [ ] **Step 6: Get a non-existent task**

```bash
curl -s http://localhost:8080/tasks/999
```
Expected: `task not found` with 404 status.

- [ ] **Step 7: Stop the server**

```bash
kill %1
```

---

### Task 8: Write tests

**Files:**
- Create: `handler_test.go`

- [ ] **Step 1: Write handler_test.go**

```go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Go convention: test file lives in the same package and is named *_test.go.
// httptest is the standard-library equivalent of Python's TestClient (Starlette).
// A table-driven test is like pytest.mark.parametrize — one test function,
// multiple input/output pairs — but written as a Go range loop over a slice
// of anonymous structs.

func newTestServer() *Server {
	return &Server{store: NewTaskStore()}
}

func TestCreateTask(t *testing.T) {
	srv := newTestServer()

	body := bytes.NewBufferString(`{"title": "learn Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.tasksHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var task Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}
	if task.Title != "learn Go" {
		t.Errorf("expected title 'learn Go', got %q", task.Title)
	}
	if task.Completed {
		t.Error("expected Completed to be false (zero value)")
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestListTasks(t *testing.T) {
	srv := newTestServer()

	// Seed two tasks directly via the store.
	srv.store.Create("first")
	srv.store.Create("second")

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()

	srv.tasksHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var tasks []Task
	json.NewDecoder(rec.Body).Decode(&tasks)

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	// Newest first: "second" should be first in the slice.
	if tasks[0].Title != "second" {
		t.Errorf("expected 'second' first, got %q", tasks[0].Title)
	}
}

func TestGetTask(t *testing.T) {
	srv := newTestServer()
	srv.store.Create("learn Go")

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rec := httptest.NewRecorder()

	srv.taskHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var task Task
	json.NewDecoder(rec.Body).Decode(&task)
	if task.Title != "learn Go" {
		t.Errorf("expected 'learn Go', got %q", task.Title)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	rec := httptest.NewRecorder()

	srv.taskHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// TestCreateTaskValidation is a table-driven test — the idiomatic Go pattern.
// Each table entry is an anonymous struct with a name, input, and expected outcome.
// The test loops over the table and runs assertions for each case.
// Compare: pytest's @pytest.mark.parametrize("name,body,want_status", [...])
func TestCreateTaskValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid task", `{"title": "learn Go"}`, http.StatusCreated},
		{"empty title", `{"title": ""}`, http.StatusBadRequest},
		{"missing title", `{}`, http.StatusBadRequest},
		{"malformed JSON", `not json`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer()

			req := httptest.NewRequest(http.MethodPost, "/tasks",
				bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			srv.tasksHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d (body: %s)",
					tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd /home/li/go-project && go test -v ./...
```
Expected: all tests PASS.

---

### Task 9: Python ↔ Go cheat sheet

**Files:**
- Create: `docs/python-go-cheatsheet.md`

- [ ] **Step 1: Write cheat sheet**

```markdown
# Python → Go Cheat Sheet

## Types

| Python | Go |
|--------|-----|
| `int` | `int` (int64 on 64-bit) |
| `str` | `string` |
| `bool` | `bool` |
| `float` | `float64` |
| `list[T]` | `[]T` (slice) |
| `dict[K,V]` | `map[K]V` |
| `tuple` | No direct equivalent (struct or multiple return) |
| `set` | No built-in (use `map[T]struct{}` or wait for stdlib) |
| `None` | `nil` (only for pointers, slices, maps, interfaces, funcs, channels) |
| `dataclass` | `struct` with exported fields |
| `Optional[T]` | `*T` (pointer, nil means absent) or `(T, bool)` |
| `Exception` | `error` interface (returned, not raised) |

## Variable declaration

```python
# Python
x = 42
name: str = "Alice"
```

```go
// Go — short declaration (inside functions only)
x := 42
name := "Alice"

// Go — explicit declaration (any scope)
var x int = 42
var name string = "Alice"
```

## Functions

```python
# Python
def add(a: int, b: int) -> int:
    return a + b
```

```go
// Go
func add(a, b int) int {
    return a + b
}
```

## Multiple return (Go's error pattern)

```python
# Python
def get_user(id: int) -> User | None:
    ...
```

```go
// Go — return value + error
func getUser(id int) (User, error) {
    ...
}
```

## Methods

```python
# Python
class Counter:
    value: int = 0
    def increment(self) -> None:
        self.value += 1
```

```go
// Go — method receiver before func name
type Counter struct {
    value int
}

func (c *Counter) Increment() {
    c.value++
}
```

## Slices (dynamic arrays)

```python
# Python
items = ["a", "b", "c"]
items.append("d")
first := items[0]
subset := items[1:3]
```

```go
// Go
items := []string{"a", "b", "c"}
items = append(items, "d")
first := items[0]
subset := items[1:3]
```

## Maps

```python
# Python
scores = {"alice": 10, "bob": 20}
scores["charlie"] = 30
val = scores.get("dave", 0)      # default
val, ok = scores["dave"]          # check existence
```

```go
// Go
scores := map[string]int{"alice": 10, "bob": 20}
scores["charlie"] = 30
val, ok := scores["dave"]         // ok is false if key missing
```

## Error handling

```python
# Python
try:
    result = do_something()
except ValueError as e:
    print(f"error: {e}")
```

```go
// Go
result, err := doSomething()
if err != nil {
    fmt.Printf("error: %v\n", err)
}
```

## Concurrency

```python
# Python (asyncio)
import asyncio
async def fetch(url: str) -> str: ...
await asyncio.gather(fetch(a), fetch(b))
```

```go
// Go (goroutines + channels)
func fetch(url string) string { ... }

// Run concurrently, collect results via channel
ch := make(chan string, 2)
go func() { ch <- fetch(a) }()
go func() { ch <- fetch(b) }()
x, y := <-ch, <-ch
```
```

---

### Task 10: Final check

- [ ] **Step 1: Run the full Go toolchain**

```bash
cd /home/li/go-project && go fmt ./... && go vet ./... && go test -v ./...
```
Expected: `go fmt` produces no diff, `go vet` reports nothing, all tests pass.

- [ ] **Step 2: See final file tree**

```
taskapi/
├── go.mod
├── main.go
├── task.go
├── store.go
├── handler.go
├── handler_test.go
└── docs/
    └── python-go-cheatsheet.md
```
