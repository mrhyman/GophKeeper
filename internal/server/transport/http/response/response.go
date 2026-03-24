package response

import (
	"encoding/json"
	"net/http"
)

// ErrorBody — стандартный формат ошибки API.
type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// JSON записывает JSON-ответ.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error записывает ошибку в стандартном формате.
func Error(w http.ResponseWriter, status int, err string, message string) {
	JSON(w, status, ErrorBody{
		Error:   err,
		Message: message,
	})
}
