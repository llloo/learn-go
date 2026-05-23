package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"taskapi/internal/handler"
	"taskapi/internal/store"
	"taskapi/internal/task"
	"testing"

	"github.com/go-chi/chi/v5"
)

// newTestServer 创建一个隔离的 Server 实例给每个测试用
// Python 类比：pytest fixture 在每个测试中创建独立环境
func newTestServer() *handler.Server {
	return &handler.Server{Store: store.NewStore()}
}

func newTestRouter() (*handler.Server, chi.Router) {
	srv := newTestServer()
	r := chi.NewRouter()
	r.Get("/tasks", srv.HandleGetTasks)
	r.Post("/tasks", srv.HandleCreateTask)
	r.Get("/tasks/{id}", srv.HandleGetTaskByID)

	return srv, r
}

func TestCreateTask(t *testing.T) {
	// Arrange — 准备请求和响应 recorder
	srv := newTestServer()

	body := strings.NewReader(`{"title": "learn Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Act — 直接调用 handler（不走真实端口）
	srv.HandleCreateTask(rec, req)

	// Assert — 验证响应
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var task task.Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// 验证各个字段
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

func TestGetTasks(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Store.Create(req.Context(), "Task 1")
	srv.Store.Create(req.Context(), "Task 2")

	srv.HandleGetTasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var tasks []task.Task
	if err := json.NewDecoder(rec.Body).Decode(&tasks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	if tasks[0].Title != "Task 2" || tasks[1].Title != "Task 1" {
		t.Errorf("expected tasks in reverse order, got %q and %q", tasks[0].Title, tasks[1].Title)
	}
	if tasks[0].ID != 2 || tasks[1].ID != 1 {
		t.Errorf("expected IDs 2 and 1, got %d and %d", tasks[0].ID, tasks[1].ID)
	}

}

func TestGetTaskByID(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rec := httptest.NewRecorder()

	srv, r := newTestRouter()
	newTask := srv.Store.Create(req.Context(), "Task 1")

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got task.Task
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.ID != newTask.ID {
		t.Errorf("expected ID %d, got %d", newTask.ID, got.ID)
	}
	if got.Title != newTask.Title {
		t.Errorf("expected title %q, got %q", newTask.Title, got.Title)
	}
	if got.Completed != newTask.Completed {
		t.Errorf("expected Completed %v, got %v", newTask.Completed, got.Completed)
	}
	if !got.CreatedAt.Equal(newTask.CreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", newTask.CreatedAt, got.CreatedAt)
	}
}

func TestGetTaskByID_NotFound(t *testing.T) {
	_, r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestGetTaskByID_InvalidID(t *testing.T) {
	_, r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
