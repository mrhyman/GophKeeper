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

func TestClient_Register(t *testing.T) {
	tests := []struct {
		name       string
		login      string
		password   string
		statusCode int
		response   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			login:      "testuser",
			password:   "testpass",
			statusCode: http.StatusOK,
			response: TokenPair{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
			wantErr: false,
		},
		{
			name:       "user already exists",
			login:      "existing",
			password:   "testpass",
			statusCode: http.StatusConflict,
			response: map[string]string{
				"error":   "conflict",
				"message": "user already exists",
			},
			wantErr: true,
		},
		{
			name:       "invalid response",
			login:      "testuser",
			password:   "testpass",
			statusCode: http.StatusOK,
			response:   "invalid json",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v1/register", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var req map[string]string
				err = json.Unmarshal(body, &req)
				require.NoError(t, err)

				assert.Equal(t, tt.login, req["login"])
				assert.Equal(t, tt.password, req["password"])

				w.WriteHeader(tt.statusCode)
				if str, ok := tt.response.(string); ok {
					w.Write([]byte(str))
				} else {
					json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := newTestClient(t, server)
			tokens, err := client.Register(tt.login, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, tokens)
				assert.Equal(t, "access-token", tokens.AccessToken)
				assert.Equal(t, "refresh-token", tokens.RefreshToken)
			}
		})
	}
}

func TestClient_Login(t *testing.T) {
	tests := []struct {
		name       string
		login      string
		password   string
		statusCode int
		response   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			login:      "testuser",
			password:   "testpass",
			statusCode: http.StatusOK,
			response: TokenPair{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
			wantErr: false,
		},
		{
			name:       "invalid credentials",
			login:      "testuser",
			password:   "wrongpass",
			statusCode: http.StatusUnauthorized,
			response: map[string]string{
				"error":   "unauthorized",
				"message": "invalid credentials",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v1/login", r.URL.Path)

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := newTestClient(t, server)
			tokens, err := client.Login(tt.login, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, tokens)
				assert.Equal(t, "access-token", tokens.AccessToken)
			}
		})
	}
}

func TestClient_Refresh(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		statusCode   int
		response     interface{}
		wantErr      bool
	}{
		{
			name:         "success",
			refreshToken: "valid-refresh-token",
			statusCode:   http.StatusOK,
			response: TokenPair{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
			},
			wantErr: false,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid-token",
			statusCode:   http.StatusUnauthorized,
			response: map[string]string{
				"error":   "unauthorized",
				"message": "invalid refresh token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v1/refresh", r.URL.Path)

				body, _ := io.ReadAll(r.Body)
				var req map[string]string
				json.Unmarshal(body, &req)
				assert.Equal(t, tt.refreshToken, req["refresh_token"])

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := newTestClient(t, server)
			tokens, err := client.Refresh(tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, tokens)
				assert.Equal(t, "new-access-token", tokens.AccessToken)
			}
		})
	}
}
