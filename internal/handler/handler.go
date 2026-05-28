package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"taskapi/internal/store"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Store store.TaskStore
}

func (s *Server) HandleGetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tasks, err := s.Store.GetAll(r.Context())

	if err != nil {
		WriteError(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		WriteError(w, "Failed to encode tasks", http.StatusInternalServerError)
		return
	}

}

func (s *Server) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var input struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(body, &input); err != nil {
		WriteError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task, err := s.Store.Create(r.Context(), input.Title)
	if err != nil {
		WriteError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(task)
	if err != nil {
		WriteError(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(resp)

}

func (s *Server) HandleGetTaskByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		WriteError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := s.Store.GetByID(r.Context(), taskID)
	if err != nil {
		WriteError(w, "Task not found", http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		WriteError(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

func (s *Server) HandlePartialUpdateTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		WriteError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var input struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
	}
	if err := json.Unmarshal(body, &input); err != nil {
		WriteError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task, err := s.Store.PartialUpdate(r.Context(), taskID, input.Title, input.Completed)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			WriteError(w, "Task not found", http.StatusNotFound)
		} else {
			WriteError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		WriteError(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

func (s *Server) HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		WriteError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	err = s.Store.Delete(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			WriteError(w, "Task not found", http.StatusNotFound)
		} else {
			WriteError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
