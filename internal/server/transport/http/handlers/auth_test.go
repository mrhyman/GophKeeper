package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	authDomain "gophkeeper/internal/server/domain/auth"
	"gophkeeper/internal/server/transport/http/dto"
	"gophkeeper/internal/server/transport/http/mocks"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	tokens := &authDomain.TokenPair{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
	}

	mockService.On("Register", mock.Anything, "testuser", "password123").
		Return(tokens, nil)

	reqBody := dto.RegisterRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "access-token-123", response.AccessToken)
	assert.Equal(t, "refresh-token-456", response.RefreshToken)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/register",
		bytes.NewBufferString(`{"login": "test"`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid JSON")
}

func TestAuthHandler_Register_EmptyLogin(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	reqBody := dto.RegisterRequest{
		Login:    "",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "login and password required")
}

func TestAuthHandler_Register_EmptyPassword(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	reqBody := dto.RegisterRequest{
		Login:    "testuser",
		Password: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "login and password required")
}

func TestAuthHandler_Register_UserExists(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	mockService.On("Register", context.Background(), "existing", "password123").
		Return(nil, authDomain.ErrUserExists)

	reqBody := dto.RegisterRequest{
		Login:    "existing",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "user already exists")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	tokens := &authDomain.TokenPair{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
	}

	mockService.On("Login", context.Background(), "testuser", "password123").
		Return(tokens, nil)

	reqBody := dto.LoginRequest{
		Login:    "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "access-token-123", response.AccessToken)
	assert.Equal(t, "refresh-token-456", response.RefreshToken)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	mockService.On("Login", context.Background(), "testuser", "wrongpassword").
		Return(nil, authDomain.ErrInvalidCredentials)

	reqBody := dto.LoginRequest{
		Login:    "testuser",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	tokens := &authDomain.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	mockService.On("Refresh", context.Background(), "old-refresh-token").
		Return(tokens, nil)

	reqBody := dto.RefreshRequest{
		RefreshToken: "old-refresh-token",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", response.AccessToken)
	assert.Equal(t, "new-refresh-token", response.RefreshToken)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := NewAuthHandler(mockService)

	mockService.On("Refresh", context.Background(), "invalid-token").
		Return(nil, authDomain.ErrInvalidToken)

	reqBody := dto.RefreshRequest{
		RefreshToken: "invalid-token",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid refresh token")

	mockService.AssertExpectations(t)
}
