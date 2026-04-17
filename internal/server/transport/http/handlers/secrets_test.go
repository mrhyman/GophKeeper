package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	secretsDomain "gophkeeper/internal/server/domain/secrets"
	"gophkeeper/internal/server/transport/http/dto"
	"gophkeeper/internal/server/transport/http/middleware"
	"gophkeeper/internal/server/transport/http/mocks"
	"gophkeeper/pkg/models"
)

func TestSecretsHandler_Create_Success(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Secret")).
		Return(nil)

	reqBody := dto.CreateSecretRequest{
		Name:     "Test Secret",
		Type:     int(models.SecretTypeLoginPassword),
		Data:     []byte("encrypted-data"),
		Metadata: map[string]string{"key": "value"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.SecretResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Test Secret", response.Name)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_Create_InvalidJSON(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets",
		bytes.NewBufferString(`{"name": "test"`))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid JSON")
}

func TestSecretsHandler_Create_ServiceError(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Secret")).
		Return(assert.AnError)

	reqBody := dto.CreateSecretRequest{
		Name:     "Test",
		Type:     int(models.SecretTypeLoginPassword),
		Data:     []byte("data"),
		Metadata: map[string]string{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/secrets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_List_Success(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	secrets := []*models.Secret{
		{
			ID:       "secret-1",
			UserID:   "user-123",
			Type:     models.SecretTypeLoginPassword,
			Name:     "Secret 1",
			Data:     []byte("data1"),
			Metadata: map[string]string{},
			Version:  1,
		},
		{
			ID:       "secret-2",
			UserID:   "user-123",
			Type:     models.SecretTypeBankCard,
			Name:     "Secret 2",
			Data:     []byte("data2"),
			Metadata: map[string]string{},
			Version:  1,
		},
	}

	mockService.On("List", mock.Anything, "user-123").
		Return(secrets, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SecretListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Secrets, 2)
	assert.Equal(t, "secret-1", response.Secrets[0].ID)
	assert.Equal(t, "secret-2", response.Secrets[1].ID)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_List_Empty(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("List", mock.Anything, "user-123").
		Return([]*models.Secret{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SecretListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Secrets, 0)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_List_ServiceError(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("List", mock.Anything, "user-123").
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_Get_Success(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	secret := &models.Secret{
		ID:       "secret-1",
		UserID:   "user-123",
		Type:     models.SecretTypeLoginPassword,
		Name:     "Test Secret",
		Data:     []byte("encrypted-data"),
		Metadata: map[string]string{"key": "value"},
		Version:  1,
	}

	mockService.On("Get", mock.Anything, "user-123", "secret-1").
		Return(secret, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/secret-1", nil)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "secret-1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SecretResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "secret-1", response.ID)
	assert.Equal(t, "Test Secret", response.Name)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_Get_NotFound(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("Get", mock.Anything, "user-123", "non-existent").
		Return(nil, secretsDomain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secrets/non-existent", nil)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "secret not found")

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_Update_Success(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("Update", mock.Anything, mock.AnythingOfType("*models.Secret")).
		Return(nil)

	reqBody := dto.UpdateSecretRequest{
		Name:     "Updated Secret",
		Type:     int(models.SecretTypeLoginPassword),
		Data:     []byte("new-encrypted-data"),
		Metadata: map[string]string{"updated": "true"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/secret-1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "secret-1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SecretResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Secret", response.Name)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_Update_NotFound(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("Update", mock.Anything, mock.AnythingOfType("*models.Secret")).
		Return(secretsDomain.ErrNotFound)

	reqBody := dto.UpdateSecretRequest{
		Name:     "Updated",
		Type:     int(models.SecretTypeLoginPassword),
		Data:     []byte("data"),
		Metadata: map[string]string{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/secrets/non-existent", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "secret not found")

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_Delete_Success(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("Delete", mock.Anything, "user-123", "secret-1").
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/secret-1", nil)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "secret-1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockService.AssertExpectations(t)
}

func TestSecretsHandler_Delete_NotFound(t *testing.T) {
	mockService := new(mocks.SecretsService)
	handler := NewSecretsHandler(mockService)

	mockService.On("Delete", mock.Anything, "user-123", "non-existent").
		Return(secretsDomain.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/non-existent", nil)

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "secret not found")

	mockService.AssertExpectations(t)
}
