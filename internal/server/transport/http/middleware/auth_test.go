package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	authDomain "gophkeeper/internal/server/domain/auth"
	"gophkeeper/internal/server/transport/http/mocks"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockService := new(mocks.AuthService)

	claims := &authDomain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		UserID: "user-123",
	}

	mockService.On("ParseAccessToken", "valid-token").
		Return(claims, nil)

	middleware := Auth(mockService)

	// Создаём тестовый handler
	nextCalled := false
	var capturedUserID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, "user-123", capturedUserID)
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	mockService := new(mocks.AuthService)
	middleware := Auth(mockService)

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestAuthMiddleware_InvalidAuthHeaderFormat(t *testing.T) {
	mockService := new(mocks.AuthService)
	middleware := Auth(mockService)

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid auth header")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockService := new(mocks.AuthService)

	mockService.On("ParseAccessToken", "invalid-token").
		Return(nil, authDomain.ErrInvalidToken)

	middleware := Auth(mockService)

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")

	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	mockService := new(mocks.AuthService)

	mockService.On("ParseAccessToken", "").
		Return(nil, authDomain.ErrInvalidToken)

	middleware := Auth(mockService)

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, nextCalled, "next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockService.AssertExpectations(t)
}

func TestGetUserID_WithValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-123")
	userID := GetUserID(ctx)
	assert.Equal(t, "user-123", userID)
}

func TestGetUserID_WithoutValue(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)
	assert.Equal(t, "", userID)
}

func TestGetUserID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, 123)
	userID := GetUserID(ctx)
	assert.Equal(t, "", userID)
}
