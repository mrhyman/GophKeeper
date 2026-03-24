package http

import (
	"encoding/json"
	"fmt"
)

// Register отправляет запрос на регистрацию.
func (c *Client) Register(login, password string) (*TokenPair, error) {
	body := map[string]string{
		"login":    login,
		"password": password,
	}

	data, err := c.doRequest("POST", "/api/v1/register", body)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	var tokens TokenPair
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("unmarshal tokens: %w", err)
	}

	return &tokens, nil
}

// Login отправляет запрос на аутентификацию.
func (c *Client) Login(login, password string) (*TokenPair, error) {
	body := map[string]string{
		"login":    login,
		"password": password,
	}

	data, err := c.doRequest("POST", "/api/v1/login", body)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	var tokens TokenPair
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("unmarshal tokens: %w", err)
	}

	return &tokens, nil
}

// Refresh обновляет пару токенов.
func (c *Client) Refresh(refreshToken string) (*TokenPair, error) {
	body := map[string]string{
		"refresh_token": refreshToken,
	}

	data, err := c.doRequest("POST", "/api/v1/refresh", body)
	if err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}

	var tokens TokenPair
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("unmarshal tokens: %w", err)
	}

	return &tokens, nil
}
