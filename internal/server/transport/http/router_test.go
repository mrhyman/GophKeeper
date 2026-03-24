package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"gophkeeper/internal/server/config"
	authDomain "gophkeeper/internal/server/domain/auth"
	syncDomain "gophkeeper/internal/server/domain/sync"
	"gophkeeper/pkg/models"
)

// Mock AuthService
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, login, password string) (*authDomain.TokenPair, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.TokenPair), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (*authDomain.TokenPair, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.TokenPair), args.Error(1)
}

func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string) (*authDomain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.TokenPair), args.Error(1)
}

func (m *mockAuthService) ParseAccessToken(tokenStr string) (*authDomain.Claims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.Claims), args.Error(1)
}

// Mock SecretsService
type mockSecretsService struct {
	mock.Mock
}

func (m *mockSecretsService) Create(ctx context.Context, secret *models.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *mockSecretsService) Update(ctx context.Context, secret *models.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *mockSecretsService) Delete(ctx context.Context, userID, secretID string) error {
	args := m.Called(ctx, userID, secretID)
	return args.Error(0)
}

func (m *mockSecretsService) Get(ctx context.Context, userID, secretID string) (*models.Secret, error) {
	args := m.Called(ctx, userID, secretID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Secret), args.Error(1)
}

func (m *mockSecretsService) List(ctx context.Context, userID string) ([]*models.Secret, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Secret), args.Error(1)
}

// Mock SyncService
type mockSyncService struct {
	mock.Mock
}

func (m *mockSyncService) Sync(ctx context.Context, userID string, req *syncDomain.SyncRequest) (*syncDomain.SyncResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*syncDomain.SyncResponse), args.Error(1)
}

func TestNewRouter(t *testing.T) {
	authSvc := new(mockAuthService)
	secretsSvc := new(mockSecretsService)
	syncSvc := new(mockSyncService)
	logger := zap.NewNop()
	jwtCfg := config.JWTConfig{}

	router := NewRouter(authSvc, secretsSvc, syncSvc, jwtCfg, logger)

	assert.NotNil(t, router)
}

func TestRouter_PublicRoutes(t *testing.T) {
	authSvc := new(mockAuthService)
	secretsSvc := new(mockSecretsService)
	syncSvc := new(mockSyncService)
	logger := zap.NewNop()
	jwtCfg := config.JWTConfig{}

	router := NewRouter(authSvc, secretsSvc, syncSvc, jwtCfg, logger)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"register", http.MethodPost, "/api/v1/register"},
		{"login", http.MethodPost, "/api/v1/login"},
		{"refresh", http.MethodPost, "/api/v1/refresh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Роут существует (не 404/405)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "route should exist")
			assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code, "method should be allowed")
		})
	}
}

func TestRouter_ProtectedRoutes_Unauthorized(t *testing.T) {
	authSvc := new(mockAuthService)
	secretsSvc := new(mockSecretsService)
	syncSvc := new(mockSyncService)
	logger := zap.NewNop()
	jwtCfg := config.JWTConfig{}

	router := NewRouter(authSvc, secretsSvc, syncSvc, jwtCfg, logger)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"list secrets", http.MethodGet, "/api/v1/secrets"},
		{"create secret", http.MethodPost, "/api/v1/secrets"},
		{"get secret", http.MethodGet, "/api/v1/secrets/123"},
		{"update secret", http.MethodPut, "/api/v1/secrets/123"},
		{"delete secret", http.MethodDelete, "/api/v1/secrets/123"},
		{"sync", http.MethodPost, "/api/v1/sync"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Без токена должен быть 401
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestRouter_SwaggerRoute(t *testing.T) {
	authSvc := new(mockAuthService)
	secretsSvc := new(mockSecretsService)
	syncSvc := new(mockSyncService)
	logger := zap.NewNop()
	jwtCfg := config.JWTConfig{}

	router := NewRouter(authSvc, secretsSvc, syncSvc, jwtCfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Swagger должен отвечать (не 404)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestRouter_ProtectedRoutes_WithValidToken(t *testing.T) {
	authSvc := new(mockAuthService)
	secretsSvc := new(mockSecretsService)
	syncSvc := new(mockSyncService)
	logger := zap.NewNop()
	jwtCfg := config.JWTConfig{}

	// Настраиваем мок для валидного токена
	authSvc.On("ParseAccessToken", "valid-token").Return(&authDomain.Claims{
		UserID: "user-123",
	}, nil)

	secretsSvc.On("List", mock.Anything, "user-123").Return([]*models.Secret{}, nil)

	router := NewRouter(authSvc, secretsSvc, syncSvc, jwtCfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	authSvc.AssertExpectations(t)
	secretsSvc.AssertExpectations(t)
}

func TestRouter_NonExistentRoute(t *testing.T) {
	authSvc := new(mockAuthService)
	secretsSvc := new(mockSecretsService)
	syncSvc := new(mockSyncService)
	logger := zap.NewNop()
	jwtCfg := config.JWTConfig{}

	router := NewRouter(authSvc, secretsSvc, syncSvc, jwtCfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
