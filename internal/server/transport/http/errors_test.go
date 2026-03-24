package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   interface{}
	}{
		{
			name:   "success response",
			status: http.StatusOK,
			body:   map[string]string{"message": "ok"},
		},
		{
			name:   "created response",
			status: http.StatusCreated,
			body:   map[string]int{"id": 123},
		},
		{
			name:   "struct response",
			status: http.StatusOK,
			body:   struct{ Name string }{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteJSON(w, tt.status, tt.body)

			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var result map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &result)
			require.NoError(t, err)
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		errCode     string
		message     string
		wantError   string
		wantMessage string
	}{
		{
			name:        "bad request",
			status:      http.StatusBadRequest,
			errCode:     "bad_request",
			message:     "invalid input",
			wantError:   "bad_request",
			wantMessage: "invalid input",
		},
		{
			name:        "unauthorized",
			status:      http.StatusUnauthorized,
			errCode:     "unauthorized",
			message:     "invalid token",
			wantError:   "unauthorized",
			wantMessage: "invalid token",
		},
		{
			name:        "not found",
			status:      http.StatusNotFound,
			errCode:     "not_found",
			message:     "resource not found",
			wantError:   "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "internal error",
			status:      http.StatusInternalServerError,
			errCode:     "internal_error",
			message:     "something went wrong",
			wantError:   "internal_error",
			wantMessage: "something went wrong",
		},
		{
			name:        "empty message",
			status:      http.StatusBadRequest,
			errCode:     "bad_request",
			message:     "",
			wantError:   "bad_request",
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteError(w, tt.status, tt.errCode, tt.message)

			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var body ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &body)
			require.NoError(t, err)

			assert.Equal(t, tt.wantError, body.Error)
			assert.Equal(t, tt.wantMessage, body.Message)
		})
	}
}

func TestErrorResponse_JSONFormat(t *testing.T) {
	w := httptest.NewRecorder()

	WriteError(w, http.StatusBadRequest, "validation_error", "field is required")

	expected := `{"error":"validation_error","message":"field is required"}`
	actual := w.Body.String()
	actual = actual[:len(actual)-1]

	assert.JSONEq(t, expected, actual)
}

func TestErrorResponse_OmitEmptyMessage(t *testing.T) {
	resp := ErrorResponse{
		Error:   "test_error",
		Message: "",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// message с omitempty должен отсутствовать
	assert.NotContains(t, string(data), `"message":""`)
}
