package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophkeeper/pkg/models"
)

func TestClient_Sync(t *testing.T) {
	tests := []struct {
		name          string
		lastSync      int64
		clientSecrets []*models.Secret
		statusCode    int
		response      interface{}
		wantErr       bool
	}{
		{
			name:          "success with empty client secrets",
			lastSync:      1234567890,
			clientSecrets: []*models.Secret{},
			statusCode:    http.StatusOK,
			response: SyncResponse{
				ServerSecrets: []*models.Secret{
					{ID: "server-1", Name: "server-secret"},
				},
				SyncTimestamp: 1234567900,
			},
			wantErr: false,
		},
		{
			name:     "success with client secrets",
			lastSync: 1234567890,
			clientSecrets: []*models.Secret{
				{ID: "client-1", Name: "client-secret"},
			},
			statusCode: http.StatusOK,
			response: SyncResponse{
				ServerSecrets: []*models.Secret{},
				SyncTimestamp: 1234567900,
			},
			wantErr: false,
		},
		{
			name:          "server error",
			lastSync:      0,
			clientSecrets: nil,
			statusCode:    http.StatusInternalServerError,
			response: map[string]string{
				"error":   "internal",
				"message": "sync failed",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v1/sync", r.URL.Path)

				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				_ = json.Unmarshal(body, &req)

				assert.Equal(t, float64(tt.lastSync), req["last_sync"])

				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := newTestClient(t, server)
			resp, err := client.Sync(tt.lastSync, tt.clientSecrets)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, int64(1234567900), resp.SyncTimestamp)
			}
		})
	}
}

func TestClient_Sync_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.Sync(0, nil)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "unmarshal")
}
