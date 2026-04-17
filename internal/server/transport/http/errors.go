package http

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse — стандартный формат ошибки API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// writeJSON записывает JSON-ответ.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError записывает ошибку в стандартном формате.
func WriteError(w http.ResponseWriter, status int, err string, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error:   err,
		Message: message,
	})
}
