package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

	_, _ = srv.Store.Create(req.Context(), "Task 1")
	_, _ = srv.Store.Create(req.Context(), "Task 2")

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

	// table driven tests

	tests := []struct {
		name           string
		taskID         string
		setup          bool
		expectedStatus int
	}{
		{"Valid ID", "1", true, http.StatusOK},
		{"Not Found", "999", false, http.StatusNotFound},
		{"Invalid ID", "abc", false, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, r := newTestRouter()
			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)
			if tt.setup {
				_, _ = srv.Store.Create(req.Context(), "Task 1")
			}

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				var task task.Task
				if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if task.ID != 1 {
					t.Errorf("expected ID 1, got %d", task.ID)
				}
				if task.Title != "Task 1" {
					t.Errorf("expected title 'Task 1', got %q", task.Title)
				}
				if task.Completed {
					t.Error("expected Completed to be false (zero value)")
				}
				if task.CreatedAt.IsZero() {
					t.Error("expected CreatedAt to be set")
				}
			}
		})
	}
}

func TestBatchCreateTasks(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedStatus int
	}{
		{"Valid Input", `{"titles": ["Task 1", "Task 2"]}`, http.StatusCreated},
		{"Empty Titles", `{"titles": []}`, http.StatusBadRequest},
		{"Invalid JSON", `{"titles": [}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer()

			body := strings.NewReader(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/tasks/batch", body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			srv.HandleBatchCreateTasks(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var results []handler.BatchResult
				if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(results) != 2 {
					t.Fatalf("expected 2 results, got %d", len(results))
				}
				for i, res := range results {
					if res.Error != "" {
						t.Errorf("unexpected error for task %d: %s", i+1, res.Error)
					} else if res.Task == nil {
						t.Errorf("expected task for result %d, got nil", i+1)
					} else if res.Task.Title != "Task "+strconv.Itoa(i+1) {
						t.Errorf("expected title 'Task %d', got %q", i+1, res.Task.Title)
					}
				}
			}
		})
	}
}
