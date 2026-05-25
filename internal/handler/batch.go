package handler

import "taskapi/internal/task"

type batchResult struct {
	Task  task.Task `json:"task"`
	Error string    `json:"error,omitempty"`
	Index int       `json:"index"`
}
