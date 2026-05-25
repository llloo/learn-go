package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"taskapi/internal/task"
	"time"
)

type BatchResult struct {
	Task  *task.Task `json:"task"`
	Error string     `json:"error,omitempty"`
	Index int        `json:"index"`
}

func (s *Server) HandleBatchCreateTasks(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var input struct {
		Titles []string `json:"titles"`
	}
	if err := json.Unmarshal(body, &input); err != nil {
		WriteError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(input.Titles) == 0 {
		WriteError(w, "titles must not be empty", http.StatusBadRequest)
		return
	}

	ch := make(chan BatchResult, len(input.Titles))
	sem := make(chan struct{}, 5) // Limit to 5 concurrent creations
	results := make([]BatchResult, len(input.Titles))

	for i, title := range input.Titles {
		go func(t string, i int) {
			sem <- struct{}{}
			defer func() { <-sem }()
			created, err := s.Store.Create(r.Context(), t)
			if err != nil {
				ch <- BatchResult{Error: err.Error(), Index: i}
				return
			}
			ch <- BatchResult{Task: &created, Index: i}
		}(title, i)
	}

	timeout := time.After(5 * time.Second)
	for i := 0; i < len(input.Titles); i++ {

		select {
		case res := <-ch:
			results[res.Index] = res
		case <-r.Context().Done():
			WriteError(w, "Request cancelled", http.StatusRequestTimeout)
			return
		case <-timeout:
			WriteError(w, "Batch processing timed out", http.StatusGatewayTimeout)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		WriteError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
