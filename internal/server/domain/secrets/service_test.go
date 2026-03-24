package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophkeeper/internal/server/repository/mocks"
	"gophkeeper/internal/server/repository/postgres"
	"gophkeeper/pkg/models"
)

func setupSecretsService(t *testing.T) (*Service, *mocks.SecretRepository) {
	secretRepo := new(mocks.SecretRepository)
	service := NewService(secretRepo)
	return service, secretRepo
}

func validSecret() *models.Secret {
	return &models.Secret{
		UserID: "user-123",
		Name:   "my-secret",
		Type:   models.SecretTypeLoginPassword,
		Data:   []byte("encrypted-data"),
	}
}

func TestService_Create_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()

	secretRepo.On("Create", ctx, secret).Return(nil)

	err := service.Create(ctx, secret)

	require.NoError(t, err)
	secretRepo.AssertExpectations(t)
}

func TestService_Create_ValidationError_EmptyName(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()
	secret.Name = ""

	err := service.Create(ctx, secret)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
	secretRepo.AssertNotCalled(t, "Create")
}

func TestService_Create_ValidationError_InvalidType(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()
	secret.Type = 0 // невалидный тип

	err := service.Create(ctx, secret)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid secret type")
	secretRepo.AssertNotCalled(t, "Create")
}

func TestService_Create_ValidationError_EmptyData(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()
	secret.Data = nil

	err := service.Create(ctx, secret)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "data is required")
	secretRepo.AssertNotCalled(t, "Create")
}

func TestService_Create_RepoError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()

	secretRepo.On("Create", ctx, secret).Return(assert.AnError)

	err := service.Create(ctx, secret)

	require.Error(t, err)
	secretRepo.AssertExpectations(t)
}

func TestService_Get_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	expected := &models.Secret{
		ID:     "secret-456",
		UserID: "user-123",
		Name:   "my-secret",
	}

	secretRepo.On("GetByID", ctx, "user-123", "secret-456").Return(expected, nil)

	secret, err := service.Get(ctx, "user-123", "secret-456")

	require.NoError(t, err)
	assert.Equal(t, expected, secret)
	secretRepo.AssertExpectations(t)
}

func TestService_Get_NotFound(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secretRepo.On("GetByID", ctx, "user-123", "nonexistent").
		Return(nil, postgres.ErrSecretNotFound)

	secret, err := service.Get(ctx, "user-123", "nonexistent")

	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, secret)
	secretRepo.AssertExpectations(t)
}

func TestService_Get_RepoError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secretRepo.On("GetByID", ctx, "user-123", "secret-456").
		Return(nil, assert.AnError)

	secret, err := service.Get(ctx, "user-123", "secret-456")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
	assert.Nil(t, secret)
	secretRepo.AssertExpectations(t)
}

func TestService_List_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	expected := []*models.Secret{
		{ID: "secret-1", Name: "first"},
		{ID: "secret-2", Name: "second"},
	}

	secretRepo.On("List", ctx, "user-123").Return(expected, nil)

	secrets, err := service.List(ctx, "user-123")

	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	secretRepo.AssertExpectations(t)
}

func TestService_List_Empty(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secretRepo.On("List", ctx, "user-123").Return([]*models.Secret{}, nil)

	secrets, err := service.List(ctx, "user-123")

	require.NoError(t, err)
	assert.Empty(t, secrets)
	secretRepo.AssertExpectations(t)
}

func TestService_List_RepoError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secretRepo.On("List", ctx, "user-123").Return(nil, assert.AnError)

	secrets, err := service.List(ctx, "user-123")

	require.Error(t, err)
	assert.Nil(t, secrets)
	secretRepo.AssertExpectations(t)
}

func TestService_Update_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()
	secret.ID = "secret-456"

	secretRepo.On("Update", ctx, secret).Return(nil)

	err := service.Update(ctx, secret)

	require.NoError(t, err)
	secretRepo.AssertExpectations(t)
}

func TestService_Update_NotFound(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()
	secret.ID = "nonexistent"

	secretRepo.On("Update", ctx, secret).Return(postgres.ErrSecretNotFound)

	err := service.Update(ctx, secret)

	require.ErrorIs(t, err, ErrNotFound)
	secretRepo.AssertExpectations(t)
}

func TestService_Update_ValidationError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secret := validSecret()
	secret.Name = "" // невалидное имя

	err := service.Update(ctx, secret)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
	secretRepo.AssertNotCalled(t, "Update")
}

func TestService_Delete_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secretRepo.On("Delete", ctx, "user-123", "secret-456").Return(nil)

	err := service.Delete(ctx, "user-123", "secret-456")

	require.NoError(t, err)
	secretRepo.AssertExpectations(t)
}

func TestService_Delete_NotFound(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secretRepo.On("Delete", ctx, "user-123", "nonexistent").
		Return(postgres.ErrSecretNotFound)

	err := service.Delete(ctx, "user-123", "nonexistent")

	require.ErrorIs(t, err, ErrNotFound)
	secretRepo.AssertExpectations(t)
}

func TestService_Delete_RepoError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	secretRepo.On("Delete", ctx, "user-123", "secret-456").
		Return(assert.AnError)

	err := service.Delete(ctx, "user-123", "secret-456")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
	secretRepo.AssertExpectations(t)
}

func TestService_Validate_AllTypes(t *testing.T) {
	service, secretRepo := setupSecretsService(t)
	ctx := context.Background()

	types := []models.SecretType{
		models.SecretTypeLoginPassword,
		models.SecretTypeTextData,
		models.SecretTypeBinaryData,
		models.SecretTypeBankCard,
	}

	for _, secretType := range types {
		t.Run(secretType.String(), func(t *testing.T) {
			secret := &models.Secret{
				Name: "test",
				Type: secretType,
				Data: []byte("data"),
			}

			secretRepo.On("Create", ctx, secret).Return(nil).Once()

			err := service.Create(ctx, secret)
			require.NoError(t, err)
		})
	}
}
