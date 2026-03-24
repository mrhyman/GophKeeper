package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateSecret(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
		secretType int
		data       []byte
		metadata   map[string]string
		statusCode int
		response   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			secretName: "my-secret",
			secretType: 1,
			data:       []byte("secret-data"),
			metadata:   map[string]string{"key": "value"},
			statusCode: http.StatusCreated,
			response: SecretResponse{
				ID:        "secret-id",
				Name:      "my-secret",
				Type:      1,
				Data:      []byte("secret-data"),
				Version:   1,
				UpdatedAt: 1234567890,
			},
			wantErr: false,
		},
		{
			name:       "server error",
			secretName: "my-secret",
			secretType: 1,
			data:       []byte("secret-data"),
			statusCode: http.StatusInternalServerError,
			response: map[string]string{
				"error":   "internal",
				"message": "database error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v1/secrets", r.URL.Path)

				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				_ = json.Unmarshal(body, &req)

				assert.Equal(t, tt.secretName, req["name"])

				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := newTestClient(t, server)
			secret, err := client.CreateSecret(tt.secretName, tt.secretType, tt.data, tt.metadata)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, secret)
				assert.Equal(t, "secret-id", secret.ID)
			}
		})
	}
}

func TestClient_GetSecret(t *testing.T) {
	tests := []struct {
		name       string
		secretID   string
		statusCode int
		response   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			secretID:   "secret-123",
			statusCode: http.StatusOK,
			response: SecretResponse{
				ID:   "secret-123",
				Name: "my-secret",
				Type: 1,
			},
			wantErr: false,
		},
		{
			name:       "not found",
			secretID:   "non-existent",
			statusCode: http.StatusNotFound,
			response: map[string]string{
				"error":   "not_found",
				"message": "secret not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/secrets/"+tt.secretID, r.URL.Path)

				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := newTestClient(t, server)
			secret, err := client.GetSecret(tt.secretID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, secret)
				assert.Equal(t, tt.secretID, secret.ID)
			}
		})
	}
}

func TestClient_ListSecrets(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "success with secrets",
			statusCode: http.StatusOK,
			response: map[string]interface{}{
				"secrets": []SecretResponse{
					{ID: "1", Name: "secret1"},
					{ID: "2", Name: "secret2"},
				},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "success empty list",
			statusCode: http.StatusOK,
			response: map[string]interface{}{
				"secrets": []SecretResponse{},
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response: map[string]string{
				"error": "unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/secrets", r.URL.Path)

				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := newTestClient(t, server)
			secrets, err := client.ListSecrets()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, secrets, tt.wantCount)
			}
		})
	}
}

func TestClient_UpdateSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/v1/secrets/secret-123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)

		assert.Equal(t, "updated-name", req["name"])

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SecretResponse{
			ID:      "secret-123",
			Name:    "updated-name",
			Version: 2,
		})
	}))
	defer server.Close()

	client := newTestClient(t, server)
	secret, err := client.UpdateSecret("secret-123", "updated-name", 1, []byte("data"), nil)

	assert.NoError(t, err)
	require.NotNil(t, secret)
	assert.Equal(t, "updated-name", secret.Name)
	assert.Equal(t, int64(2), secret.Version)
}

func TestClient_DeleteSecret(t *testing.T) {
	tests := []struct {
		name       string
		secretID   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "success",
			secretID:   "secret-123",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "not found",
			secretID:   "non-existent",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Equal(t, "/api/v1/secrets/"+tt.secretID, r.URL.Path)

				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := newTestClient(t, server)
			err := client.DeleteSecret(tt.secretID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
