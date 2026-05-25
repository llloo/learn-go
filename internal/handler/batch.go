package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"taskapi/internal/task"
)

type batchResult struct {
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

	ch := make(chan batchResult, len(input.Titles))
	results := make([]batchResult, len(input.Titles))
	for i, title := range input.Titles {
		go func(t string, i int) {
			created, err := s.Store.Create(r.Context(), t)
			if err != nil {
				ch <- batchResult{Error: err.Error(), Index: i}
				return
			}
			ch <- batchResult{Task: &created, Index: i}
		}(title, i)
	}

	for i := 0; i < len(input.Titles); i++ {
		result := <-ch
		results[result.Index] = result
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(results)
}
