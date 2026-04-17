package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gophkeeper/internal/server/repository/mocks"
	"gophkeeper/internal/server/repository/postgres"
	"gophkeeper/pkg/models"
)

func setupSyncService(t *testing.T) (*Service, *mocks.SecretRepository) {
	secretRepo := new(mocks.SecretRepository)
	service := NewService(secretRepo)
	return service, secretRepo
}

func TestService_Sync_NoChanges(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{},
	}

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return([]*models.Secret{}, nil)

	resp, err := service.Sync(ctx, "user-123", req)

	require.NoError(t, err)
	assert.Empty(t, resp.ServerSecrets)
	assert.NotZero(t, resp.SyncTimestamp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_ServerChangesOnly(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{},
	}

	serverSecrets := []*models.Secret{
		{ID: "secret-1", Name: "server-secret", Version: 2},
	}

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return(serverSecrets, nil)

	resp, err := service.Sync(ctx, "user-123", req)

	require.NoError(t, err)
	assert.Len(t, resp.ServerSecrets, 1)
	assert.Equal(t, "secret-1", resp.ServerSecrets[0].ID)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_ClientCreatesNew(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	clientSecret := &models.Secret{
		ID:      "new-secret",
		Name:    "from-client",
		Version: 1,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	// Секрет не существует на сервере
	secretRepo.On("GetByID", ctx, "user-123", "new-secret").
		Return(nil, postgres.ErrSecretNotFound)

	// Создаём новый
	secretRepo.On("Create", ctx, mock.MatchedBy(func(s *models.Secret) bool {
		return s.ID == "new-secret" && s.UserID == "user-123"
	})).Return(nil)

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return([]*models.Secret{}, nil)

	resp, err := service.Sync(ctx, "user-123", req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_ClientUpdatesExisting(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	clientSecret := &models.Secret{
		ID:      "existing-secret",
		Name:    "updated-name",
		Version: 3,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	// Существующий секрет с меньшей версией
	existingSecret := &models.Secret{
		ID:      "existing-secret",
		Name:    "old-name",
		Version: 2,
	}

	secretRepo.On("GetByID", ctx, "user-123", "existing-secret").
		Return(existingSecret, nil)

	secretRepo.On("Update", ctx, mock.MatchedBy(func(s *models.Secret) bool {
		return s.ID == "existing-secret" && s.Version == 3
	})).Return(nil)

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return([]*models.Secret{}, nil)

	resp, err := service.Sync(ctx, "user-123", req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_ClientDeletesExisting(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	clientSecret := &models.Secret{
		ID:        "to-delete",
		Version:   3,
		IsDeleted: true,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	existingSecret := &models.Secret{
		ID:      "to-delete",
		Version: 2,
	}

	secretRepo.On("GetByID", ctx, "user-123", "to-delete").
		Return(existingSecret, nil)

	secretRepo.On("Delete", ctx, "user-123", "to-delete").Return(nil)

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return([]*models.Secret{}, nil)

	resp, err := service.Sync(ctx, "user-123", req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_ServerWinsConflict(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	// Клиент отправляет версию 2
	clientSecret := &models.Secret{
		ID:      "conflict-secret",
		Name:    "client-version",
		Version: 2,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	// На сервере уже версия 3 — конфликт, сервер побеждает
	existingSecret := &models.Secret{
		ID:      "conflict-secret",
		Name:    "server-version",
		Version: 3,
	}

	secretRepo.On("GetByID", ctx, "user-123", "conflict-secret").
		Return(existingSecret, nil)

	// Update НЕ вызывается, т.к. версия клиента меньше

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return([]*models.Secret{existingSecret}, nil)

	resp, err := service.Sync(ctx, "user-123", req)

	require.NoError(t, err)
	assert.Len(t, resp.ServerSecrets, 1)
	assert.Equal(t, "server-version", resp.ServerSecrets[0].Name)
	secretRepo.AssertNotCalled(t, "Update")
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_GetModifiedAfterError(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{},
	}

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return(nil, assert.AnError)

	resp, err := service.Sync(ctx, "user-123", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_CreateError(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	clientSecret := &models.Secret{
		ID:      "new-secret",
		Name:    "test",
		Version: 1,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	secretRepo.On("GetByID", ctx, "user-123", "new-secret").
		Return(nil, postgres.ErrSecretNotFound)

	secretRepo.On("Create", ctx, mock.Anything).Return(assert.AnError)

	resp, err := service.Sync(ctx, "user-123", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_UpdateError(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	clientSecret := &models.Secret{
		ID:      "existing-secret",
		Name:    "updated",
		Version: 3,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	existingSecret := &models.Secret{
		ID:      "existing-secret",
		Version: 2,
	}

	secretRepo.On("GetByID", ctx, "user-123", "existing-secret").
		Return(existingSecret, nil)

	secretRepo.On("Update", ctx, mock.Anything).Return(assert.AnError)

	resp, err := service.Sync(ctx, "user-123", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_DeleteError(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	clientSecret := &models.Secret{
		ID:        "to-delete",
		Version:   3,
		IsDeleted: true,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	existingSecret := &models.Secret{
		ID:      "to-delete",
		Version: 2,
	}

	secretRepo.On("GetByID", ctx, "user-123", "to-delete").
		Return(existingSecret, nil)

	secretRepo.On("Delete", ctx, "user-123", "to-delete").
		Return(assert.AnError)

	resp, err := service.Sync(ctx, "user-123", req)

	require.Error(t, err)
	assert.Nil(t, resp)
	secretRepo.AssertExpectations(t)
}

func TestService_Sync_SetsUserID(t *testing.T) {
	service, secretRepo := setupSyncService(t)
	ctx := context.Background()

	// Клиент может отправить секрет без UserID или с чужим UserID
	clientSecret := &models.Secret{
		ID:      "new-secret",
		UserID:  "attacker-id", // попытка подмены
		Name:    "test",
		Version: 1,
	}

	req := &SyncRequest{
		LastSync:      1000,
		ClientSecrets: []*models.Secret{clientSecret},
	}

	secretRepo.On("GetByID", ctx, "user-123", "new-secret").
		Return(nil, postgres.ErrSecretNotFound)

	// Проверяем, что UserID перезаписан на правильный
	secretRepo.On("Create", ctx, mock.MatchedBy(func(s *models.Secret) bool {
		return s.UserID == "user-123"
	})).Return(nil)

	secretRepo.On("GetModifiedAfter", ctx, "user-123", int64(1000)).
		Return([]*models.Secret{}, nil)

	resp, err := service.Sync(ctx, "user-123", req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	secretRepo.AssertExpectations(t)
}
