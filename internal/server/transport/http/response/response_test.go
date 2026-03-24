package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       interface{}
		wantStatus int
	}{
		{
			name:       "success response",
			status:     http.StatusOK,
			body:       map[string]string{"message": "ok"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "created response",
			status:     http.StatusCreated,
			body:       map[string]int{"id": 123},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "struct response",
			status:     http.StatusOK,
			body:       struct{ Name string }{"test"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			JSON(w, tt.status, tt.body)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var result map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &result)
			require.NoError(t, err)
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		errCode     string
		message     string
		wantStatus  int
		wantError   string
		wantMessage string
	}{
		{
			name:        "bad request",
			status:      http.StatusBadRequest,
			errCode:     "bad_request",
			message:     "invalid input",
			wantStatus:  http.StatusBadRequest,
			wantError:   "bad_request",
			wantMessage: "invalid input",
		},
		{
			name:        "unauthorized",
			status:      http.StatusUnauthorized,
			errCode:     "unauthorized",
			message:     "invalid token",
			wantStatus:  http.StatusUnauthorized,
			wantError:   "unauthorized",
			wantMessage: "invalid token",
		},
		{
			name:        "not found",
			status:      http.StatusNotFound,
			errCode:     "not_found",
			message:     "resource not found",
			wantStatus:  http.StatusNotFound,
			wantError:   "not_found",
			wantMessage: "resource not found",
		},
		{
			name:        "internal error",
			status:      http.StatusInternalServerError,
			errCode:     "internal_error",
			message:     "something went wrong",
			wantStatus:  http.StatusInternalServerError,
			wantError:   "internal_error",
			wantMessage: "something went wrong",
		},
		{
			name:        "empty message",
			status:      http.StatusBadRequest,
			errCode:     "bad_request",
			message:     "",
			wantStatus:  http.StatusBadRequest,
			wantError:   "bad_request",
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			Error(w, tt.status, tt.errCode, tt.message)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var body ErrorBody
			err := json.Unmarshal(w.Body.Bytes(), &body)
			require.NoError(t, err)

			assert.Equal(t, tt.wantError, body.Error)
			assert.Equal(t, tt.wantMessage, body.Message)
		})
	}
}

func TestErrorBody_JSONFormat(t *testing.T) {
	w := httptest.NewRecorder()

	Error(w, http.StatusBadRequest, "validation_error", "field is required")

	expected := `{"error":"validation_error","message":"field is required"}`
	// Убираем trailing newline от json.Encoder
	actual := w.Body.String()
	actual = actual[:len(actual)-1]

	assert.JSONEq(t, expected, actual)
}
