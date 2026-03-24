package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	syncDomain "gophkeeper/internal/server/domain/sync"
	"gophkeeper/internal/server/transport/http/dto"
	"gophkeeper/internal/server/transport/http/middleware"
	"gophkeeper/internal/server/transport/http/mocks"
	"gophkeeper/pkg/models"
)

func TestSyncHandler_Sync_Success(t *testing.T) {
	mockService := new(mocks.SyncService)
	handler := NewSyncHandler(mockService)

	now := time.Now().Unix()
	lastSync := now - 3600

	serverSecrets := []*models.Secret{
		{
			ID:        "secret-1",
			UserID:    "user-123",
			Type:      models.SecretTypeLoginPassword,
			Name:      "Server Secret 1",
			Data:      []byte("data1"),
			Metadata:  map[string]string{},
			Version:   2,
			UpdatedAt: now,
			IsDeleted: false,
		},
		{
			ID:        "secret-2",
			UserID:    "user-123",
			Type:      models.SecretTypeBankCard,
			Name:      "Server Secret 2",
			Data:      []byte("data2"),
			Metadata:  map[string]string{},
			Version:   1,
			UpdatedAt: now,
			IsDeleted: false,
		},
	}

	syncResponse := &syncDomain.SyncResponse{
		ServerSecrets: serverSecrets,
		SyncTimestamp: now,
	}

	mockService.On("Sync", mock.Anything, "user-123", mock.AnythingOfType("*sync.SyncRequest")).
		Return(syncResponse, nil)

	reqBody := dto.SyncRequest{
		LastSync: lastSync,
		ClientSecrets: []dto.SecretResponse{
			{
				ID:        "client-secret-1",
				Name:      "Client Secret",
				Type:      int(models.SecretTypeTextData),
				Data:      []byte("client-data"),
				Metadata:  map[string]string{},
				Version:   1,
				UpdatedAt: now - 1800,
				IsDeleted: false,
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Sync(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SyncResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.ServerSecrets, 2)
	assert.Equal(t, now, response.SyncTimestamp)
	assert.Equal(t, "secret-1", response.ServerSecrets[0].ID)
	assert.Equal(t, "secret-2", response.ServerSecrets[1].ID)

	mockService.AssertExpectations(t)
}

func TestSyncHandler_Sync_EmptyClientSecrets(t *testing.T) {
	mockService := new(mocks.SyncService)
	handler := NewSyncHandler(mockService)

	now := time.Now().Unix()
	lastSync := now - 3600

	serverSecrets := []*models.Secret{
		{
			ID:        "secret-1",
			UserID:    "user-123",
			Type:      models.SecretTypeLoginPassword,
			Name:      "Server Secret",
			Data:      []byte("data"),
			Metadata:  map[string]string{},
			Version:   1,
			UpdatedAt: now,
		},
	}

	syncResponse := &syncDomain.SyncResponse{
		ServerSecrets: serverSecrets,
		SyncTimestamp: now,
	}

	mockService.On("Sync", mock.Anything, "user-123", mock.AnythingOfType("*sync.SyncRequest")).
		Return(syncResponse, nil)

	reqBody := dto.SyncRequest{
		LastSync:      lastSync,
		ClientSecrets: []dto.SecretResponse{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Sync(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SyncResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.ServerSecrets, 1)

	mockService.AssertExpectations(t)
}

func TestSyncHandler_Sync_InvalidJSON(t *testing.T) {
	mockService := new(mocks.SyncService)
	handler := NewSyncHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync",
		bytes.NewBufferString(`{"last_sync": `))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Sync(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid JSON")
}

func TestSyncHandler_Sync_ServiceError(t *testing.T) {
	mockService := new(mocks.SyncService)
	handler := NewSyncHandler(mockService)

	mockService.On("Sync", mock.Anything, "user-123", mock.AnythingOfType("*sync.SyncRequest")).
		Return(nil, assert.AnError)

	reqBody := dto.SyncRequest{
		LastSync:      time.Now().Unix() - 3600,
		ClientSecrets: []dto.SecretResponse{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Sync(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestSyncHandler_Sync_WithDeletedSecrets(t *testing.T) {
	mockService := new(mocks.SyncService)
	handler := NewSyncHandler(mockService)

	now := time.Now().Unix()

	serverSecrets := []*models.Secret{
		{
			ID:        "secret-1",
			UserID:    "user-123",
			Type:      models.SecretTypeLoginPassword,
			Name:      "Active Secret",
			Data:      []byte("data1"),
			Version:   1,
			UpdatedAt: now,
			IsDeleted: false,
		},
		{
			ID:        "secret-2",
			UserID:    "user-123",
			Type:      models.SecretTypeBankCard,
			Name:      "Deleted Secret",
			Data:      []byte("data2"),
			Version:   2,
			UpdatedAt: now,
			IsDeleted: true,
		},
	}

	syncResponse := &syncDomain.SyncResponse{
		ServerSecrets: serverSecrets,
		SyncTimestamp: now,
	}

	mockService.On("Sync", mock.Anything, "user-123", mock.AnythingOfType("*sync.SyncRequest")).
		Return(syncResponse, nil)

	reqBody := dto.SyncRequest{
		LastSync: now - 3600,
		ClientSecrets: []dto.SecretResponse{
			{
				ID:        "secret-2",
				Name:      "Deleted Secret",
				Type:      int(models.SecretTypeBankCard),
				Data:      []byte("data2"),
				Version:   1,
				UpdatedAt: now - 1800,
				IsDeleted: false,
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.Sync(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SyncResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.ServerSecrets, 2)

	var deletedSecret *dto.SecretResponse
	for i := range response.ServerSecrets {
		if response.ServerSecrets[i].ID == "secret-2" {
			deletedSecret = &response.ServerSecrets[i]
			break
		}
	}
	require.NotNil(t, deletedSecret)
	assert.True(t, deletedSecret.IsDeleted)

	mockService.AssertExpectations(t)
}
