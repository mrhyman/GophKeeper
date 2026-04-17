package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repoMocks "gophkeeper/internal/client/repository/mocks"
	"gophkeeper/internal/client/transport/http"
	httpMocks "gophkeeper/internal/client/transport/mocks"
	"gophkeeper/pkg/models"
)

func setupSyncService(t *testing.T) (*Service, *repoMocks.SecretRepository, *repoMocks.AuthRepository, *httpMocks.HTTPClient) {
	secretRepo := new(repoMocks.SecretRepository)
	authRepo := new(repoMocks.AuthRepository)
	httpClient := new(httpMocks.HTTPClient)
	service := NewService(secretRepo, authRepo, httpClient)
	return service, secretRepo, authRepo, httpClient
}

func TestService_Sync_NoChanges(t *testing.T) {
	service, secretRepo, authRepo, httpClient := setupSyncService(t)

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return([]*models.Secret{}, nil)

	httpClient.On("Sync", int64(1000), []*models.Secret(nil)).
		Return(&http.SyncResponse{
			ServerSecrets: []*models.Secret{},
			SyncTimestamp: 2000,
		}, nil)

	authRepo.On("SaveLastSync", int64(2000)).Return(nil)

	result, err := service.Sync()

	require.NoError(t, err)
	assert.Equal(t, 0, result.Uploaded)
	assert.Equal(t, 0, result.Downloaded)
}

func TestService_Sync_UploadOnly(t *testing.T) {
	service, secretRepo, authRepo, httpClient := setupSyncService(t)

	localSecrets := []*models.Secret{
		{ID: "local-1", Name: "new-secret", UpdatedAt: 1500},
	}

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return(localSecrets, nil)

	httpClient.On("Sync", int64(1000), mock.MatchedBy(func(s []*models.Secret) bool {
		return len(s) == 1 && s[0].ID == "local-1"
	})).Return(&http.SyncResponse{
		ServerSecrets: []*models.Secret{},
		SyncTimestamp: 2000,
	}, nil)

	authRepo.On("SaveLastSync", int64(2000)).Return(nil)

	result, err := service.Sync()

	require.NoError(t, err)
	assert.Equal(t, 1, result.Uploaded)
	assert.Equal(t, 0, result.Downloaded)
}

func TestService_Sync_DownloadOnly(t *testing.T) {
	service, secretRepo, authRepo, httpClient := setupSyncService(t)

	serverSecrets := []*models.Secret{
		{ID: "server-1", Name: "from-server"},
	}

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return([]*models.Secret{}, nil)

	httpClient.On("Sync", int64(1000), []*models.Secret(nil)).
		Return(&http.SyncResponse{
			ServerSecrets: serverSecrets,
			SyncTimestamp: 2000,
		}, nil)

	secretRepo.On("UpsertBatch", serverSecrets).Return(nil)
	authRepo.On("SaveLastSync", int64(2000)).Return(nil)

	result, err := service.Sync()

	require.NoError(t, err)
	assert.Equal(t, 0, result.Uploaded)
	assert.Equal(t, 1, result.Downloaded)
}

func TestService_Sync_BothDirections(t *testing.T) {
	service, secretRepo, authRepo, httpClient := setupSyncService(t)

	localSecrets := []*models.Secret{
		{ID: "local-1", UpdatedAt: 1500},
		{ID: "local-2", UpdatedAt: 500},
	}

	serverSecrets := []*models.Secret{
		{ID: "server-1"},
		{ID: "server-2"},
	}

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return(localSecrets, nil)

	httpClient.On("Sync", int64(1000), mock.MatchedBy(func(s []*models.Secret) bool {
		return len(s) == 1 && s[0].ID == "local-1"
	})).Return(&http.SyncResponse{
		ServerSecrets: serverSecrets,
		SyncTimestamp: 2000,
	}, nil)

	secretRepo.On("UpsertBatch", serverSecrets).Return(nil)
	authRepo.On("SaveLastSync", int64(2000)).Return(nil)

	result, err := service.Sync()

	require.NoError(t, err)
	assert.Equal(t, 1, result.Uploaded)
	assert.Equal(t, 2, result.Downloaded)
}

func TestService_Sync_GetLastSyncError(t *testing.T) {
	service, _, authRepo, _ := setupSyncService(t)

	authRepo.On("GetLastSync").Return(int64(0), assert.AnError)

	result, err := service.Sync()

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestService_Sync_ListError(t *testing.T) {
	service, secretRepo, authRepo, _ := setupSyncService(t)

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return(nil, assert.AnError)

	result, err := service.Sync()

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestService_Sync_HTTPError(t *testing.T) {
	service, secretRepo, authRepo, httpClient := setupSyncService(t)

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return([]*models.Secret{}, nil)
	httpClient.On("Sync", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	result, err := service.Sync()

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestService_Sync_UpsertBatchError(t *testing.T) {
	service, secretRepo, authRepo, httpClient := setupSyncService(t)

	serverSecrets := []*models.Secret{{ID: "server-1"}}

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return([]*models.Secret{}, nil)
	httpClient.On("Sync", mock.Anything, mock.Anything).
		Return(&http.SyncResponse{
			ServerSecrets: serverSecrets,
			SyncTimestamp: 2000,
		}, nil)
	secretRepo.On("UpsertBatch", serverSecrets).Return(assert.AnError)

	result, err := service.Sync()

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestService_Sync_SaveLastSyncError(t *testing.T) {
	service, secretRepo, authRepo, httpClient := setupSyncService(t)

	authRepo.On("GetLastSync").Return(int64(1000), nil)
	secretRepo.On("List").Return([]*models.Secret{}, nil)
	httpClient.On("Sync", mock.Anything, mock.Anything).
		Return(&http.SyncResponse{
			ServerSecrets: []*models.Secret{},
			SyncTimestamp: 2000,
		}, nil)
	authRepo.On("SaveLastSync", int64(2000)).Return(assert.AnError)

	result, err := service.Sync()

	require.Error(t, err)
	assert.Nil(t, result)
}
