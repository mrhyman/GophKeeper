package http

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"

	"gophkeeper/pkg/models"
)

// Sync отправляет запрос на синхронизацию.
func (c *Client) Sync(lastSync int64, clientSecrets []*models.Secret) (*SyncResponse, error) {
	body := map[string]interface{}{
		"last_sync":      lastSync,
		"client_secrets": clientSecrets,
	}

	data, err := c.doRequest(nethttp.MethodPost, "/api/v1/sync", body)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	var resp SyncResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal sync response: %w", err)
	}

	return &resp, nil
}
