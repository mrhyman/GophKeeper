package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gophkeeper/pkg/models"
)

// TokenPair содержит пару токенов.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// SyncResponse содержит ответ сервера на синхронизацию.
type SyncResponse struct {
	ServerSecrets []*models.Secret `json:"secrets"`
	SyncTimestamp int64            `json:"sync_timestamp"`
}

// Client — HTTP-клиент для взаимодействия с сервером GophKeeper.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	accessToken string
}

// NewClient создаёт новый HTTP-клиент.
func NewClient(baseURL string, caFile string) (*Client, error) {
	tlsConfig, err := buildTLSConfig(caFile)
	if err != nil {
		return nil, fmt.Errorf("tls config: %w", err)
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}, nil
}

func buildTLSConfig(caFile string) (*tls.Config, error) {
	if caFile == "" {
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}, nil
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA file: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}, nil
}

// SetAccessToken устанавливает токен для авторизованных запросов.
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

// Ping проверяет доступность сервера и валидность токена.
func (c *Client) Ping() error {
	_, err := c.doRequest(http.MethodGet, "/api/v1/secrets", nil)
	return err
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("server error %d: %s — %s",
			resp.StatusCode, errResp.Error, errResp.Message)
	}

	return respBody, nil
}
