package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"taskapi/internal/handler"
	"testing"
)

func BenchmarkCreateTask(b *testing.B) {
	srv := newTestServer()

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		title := "task " + strconv.Itoa(i)
		body := strings.NewReader(`{"title": "` + title + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/tasks", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		srv.HandleCreateTask(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("expected status 201, got %d", rec.Code)
		}
	}
}

func BenchmarkBatchCreateTasks_Concurrent(b *testing.B) {
	srv := newTestServer()
	titles := make([]string, 100)
	for j := 0; j < 100; j++ {
		titles[j] = "task " + strconv.Itoa(j+1)
	}

	b.ResetTimer()

	for b.Loop() {
		ch := make(chan handler.BatchResult, len(titles))
		for i, title := range titles {
			go func(t string, idx int) {
				created, err := srv.Store.Create(context.Background(), t)
				if err != nil {
					ch <- handler.BatchResult{Error: err.Error(), Index: idx}
					return
				}
				ch <- handler.BatchResult{Task: &created, Index: idx}
			}(title, i)
		}
		for range titles {
			<-ch
		}
	}
}

func BenchmarkBatchCreateTasks_Serial(b *testing.B) {
	srv := newTestServer()
	titles := make([]string, 100)
	for j := 0; j < 100; j++ {
		titles[j] = "task " + strconv.Itoa(j+1)
	}

	b.ResetTimer()

	for b.Loop() {
		handleBatchCreateSerial(srv, titles)
	}
}

func handleBatchCreateSerial(srv *handler.Server, titles []string) []handler.BatchResult {
	results := make([]handler.BatchResult, len(titles))
	for i, title := range titles {
		task, err := srv.Store.Create(context.Background(), title)
		if err != nil {
			results[i] = handler.BatchResult{Error: err.Error(), Index: i}
		} else {
			results[i] = handler.BatchResult{Task: &task, Index: i}
		}
	}
	return results
}
