package handler

import (
	"encoding/json"
	"net/http"
)

func (s *Server) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Token string `json:"token"`
	}

	jwtStr, err := s.Auth.GenerateToken("user123")
	if err != nil {
		WriteError(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Token = jwtStr

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		WriteError(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
