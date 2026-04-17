package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SecretResponse — ответ сервера с данными секрета.
type SecretResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      int               `json:"type"`
	Data      []byte            `json:"data"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Version   int64             `json:"version"`
	UpdatedAt int64             `json:"updated_at"`
	IsDeleted bool              `json:"is_deleted,omitempty"`
}

// CreateSecret отправляет запрос на создание секрета.
func (c *Client) CreateSecret(name string, secretType int, data []byte, metadata map[string]string) (*SecretResponse, error) {
	body := map[string]interface{}{
		"name":     name,
		"type":     secretType,
		"data":     data,
		"metadata": metadata,
	}

	respData, err := c.doRequest(http.MethodPost, "/api/v1/secrets", body)
	if err != nil {
		return nil, fmt.Errorf("create secret: %w", err)
	}

	var secret SecretResponse
	if err := json.Unmarshal(respData, &secret); err != nil {
		return nil, fmt.Errorf("unmarshal secret: %w", err)
	}

	return &secret, nil
}

// GetSecret получает секрет по ID.
func (c *Client) GetSecret(id string) (*SecretResponse, error) {
	data, err := c.doRequest(http.MethodGet, "/api/v1/secrets/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("get secret: %w", err)
	}

	var secret SecretResponse
	if err := json.Unmarshal(data, &secret); err != nil {
		return nil, fmt.Errorf("unmarshal secret: %w", err)
	}

	return &secret, nil
}

// ListSecrets получает список секретов.
func (c *Client) ListSecrets() ([]SecretResponse, error) {
	data, err := c.doRequest(http.MethodGet, "/api/v1/secrets", nil)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	var resp struct {
		Secrets []SecretResponse `json:"secrets"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal secrets: %w", err)
	}

	return resp.Secrets, nil
}

// UpdateSecret обновляет секрет.
func (c *Client) UpdateSecret(id, name string, secretType int, secretData []byte, metadata map[string]string) (*SecretResponse, error) {
	body := map[string]interface{}{
		"name":     name,
		"type":     secretType,
		"data":     secretData,
		"metadata": metadata,
	}

	respData, err := c.doRequest(http.MethodPut, "/api/v1/secrets/"+id, body)
	if err != nil {
		return nil, fmt.Errorf("update secret: %w", err)
	}

	var secret SecretResponse
	if err := json.Unmarshal(respData, &secret); err != nil {
		return nil, fmt.Errorf("unmarshal secret: %w", err)
	}

	return &secret, nil
}

// DeleteSecret удаляет секрет.
func (c *Client) DeleteSecret(id string) error {
	_, err := c.doRequest(http.MethodDelete, "/api/v1/secrets/"+id, nil)
	return err
}
