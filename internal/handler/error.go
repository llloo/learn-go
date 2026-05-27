package handler

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIError{Message: message})
}
