package http

import (
	"encoding/json"
	"net/http"
)

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, ErrorResponse{Error: msg, Status: status})
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
