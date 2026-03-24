package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
	return client
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		caFile  string
		wantErr bool
	}{
		{
			name:    "without CA file",
			baseURL: "https://localhost:8443",
			caFile:  "",
			wantErr: false,
		},
		{
			name:    "with non-existent CA file",
			baseURL: "https://localhost:8443",
			caFile:  "/non/existent/ca.crt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, tt.caFile)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_SetAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	client := newTestClient(t, server)

	assert.Empty(t, client.accessToken)
	client.SetAccessToken("test-token")
	assert.Equal(t, "test-token", client.accessToken)
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			response:   `{"secrets":[]}`,
			wantErr:    false,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"error":"unauthorized","message":"invalid token"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/secrets", r.URL.Path)

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := newTestClient(t, server)
			err := client.Ping()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_DoRequest_WithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-token", authHeader)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"secrets":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	client.SetAccessToken("test-token")

	err := client.Ping()
	assert.NoError(t, err)
}
