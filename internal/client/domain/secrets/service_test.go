package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gophkeeper/internal/client/repository/bolt"
	repoMocks "gophkeeper/internal/client/repository/mocks"
	"gophkeeper/pkg/crypto"
	"gophkeeper/pkg/models"
)

const testMasterPassword = "test-master-password"

func setupSecretsService(t *testing.T) (*Service, *repoMocks.SecretRepository) {
	secretRepo := new(repoMocks.SecretRepository)
	service := NewService(secretRepo, testMasterPassword)
	return service, secretRepo
}

func TestService_Create_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	payload := &models.LoginPassword{
		Login:    "user@example.com",
		Password: "secret123",
	}

	secretRepo.On("Create", mock.MatchedBy(func(s *models.Secret) bool {
		return s.Name == "my-login" &&
			s.Type == models.SecretTypeLoginPassword &&
			len(s.Data) > 0 &&
			s.Version == 1
	})).Return(nil)

	secret, err := service.Create("my-login", models.SecretTypeLoginPassword, payload, nil)

	require.NoError(t, err)
	assert.NotEmpty(t, secret.ID)
	assert.Equal(t, "my-login", secret.Name)
	assert.Equal(t, models.SecretTypeLoginPassword, secret.Type)
	assert.NotEmpty(t, secret.Data)
	secretRepo.AssertExpectations(t)
}

func TestService_Create_WithMetadata(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	payload := &models.TextData{Text: "some text"}
	metadata := map[string]string{"category": "work"}

	secretRepo.On("Create", mock.MatchedBy(func(s *models.Secret) bool {
		return s.Metadata["category"] == "work"
	})).Return(nil)

	secret, err := service.Create("note", models.SecretTypeTextData, payload, metadata)

	require.NoError(t, err)
	assert.Equal(t, "work", secret.Metadata["category"])
	secretRepo.AssertExpectations(t)
}

func TestService_Create_RepoError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	payload := &models.TextData{Text: "test"}

	secretRepo.On("Create", mock.Anything).Return(assert.AnError)

	secret, err := service.Create("test", models.SecretTypeTextData, payload, nil)

	require.Error(t, err)
	assert.Nil(t, secret)
}

func TestService_Get_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	expected := &models.Secret{
		ID:   "secret-123",
		Name: "my-secret",
	}

	secretRepo.On("GetByID", "secret-123").Return(expected, nil)

	secret, err := service.Get("secret-123")

	require.NoError(t, err)
	assert.Equal(t, expected, secret)
	secretRepo.AssertExpectations(t)
}

func TestService_Get_NotFound(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	secretRepo.On("GetByID", "nonexistent").Return(nil, bolt.ErrSecretNotFound)

	secret, err := service.Get("nonexistent")

	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, secret)
}

func TestService_Decrypt_Success(t *testing.T) {
	service, _ := setupSecretsService(t)

	original := &models.LoginPassword{
		Login:    "user@example.com",
		Password: "secret123",
		URI:      "https://example.com",
	}

	// Шифруем данные
	encrypted, err := crypto.Encrypt([]byte(`{"login":"user@example.com","password":"secret123","uri":"https://example.com"}`), testMasterPassword)
	require.NoError(t, err)

	secret := &models.Secret{
		ID:   "secret-123",
		Data: encrypted,
	}

	var decrypted models.LoginPassword
	err = service.Decrypt(secret, &decrypted)

	require.NoError(t, err)
	assert.Equal(t, original.Login, decrypted.Login)
	assert.Equal(t, original.Password, decrypted.Password)
	assert.Equal(t, original.URI, decrypted.URI)
}

func TestService_Decrypt_WrongPassword(t *testing.T) {
	// Сервис с другим паролем
	secretRepo := new(repoMocks.SecretRepository)
	service := NewService(secretRepo, "wrong-password")

	encrypted, _ := crypto.Encrypt([]byte(`{"text":"secret"}`), testMasterPassword)

	secret := &models.Secret{
		ID:   "secret-123",
		Data: encrypted,
	}

	var decrypted models.TextData
	err := service.Decrypt(secret, &decrypted)

	require.Error(t, err)
}

func TestService_List_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	expected := []*models.Secret{
		{ID: "secret-1", Name: "first"},
		{ID: "secret-2", Name: "second"},
	}

	secretRepo.On("List").Return(expected, nil)

	secrets, err := service.List()

	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	secretRepo.AssertExpectations(t)
}

func TestService_List_Empty(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	secretRepo.On("List").Return([]*models.Secret{}, nil)

	secrets, err := service.List()

	require.NoError(t, err)
	assert.Empty(t, secrets)
}

func TestService_List_Error(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	secretRepo.On("List").Return(nil, assert.AnError)

	secrets, err := service.List()

	require.Error(t, err)
	assert.Nil(t, secrets)
}

func TestService_Update_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	existing := &models.Secret{
		ID:      "secret-123",
		Name:    "old-name",
		Type:    models.SecretTypeLoginPassword,
		Version: 1,
	}

	secretRepo.On("GetByID", "secret-123").Return(existing, nil)
	secretRepo.On("Update", mock.MatchedBy(func(s *models.Secret) bool {
		return s.Name == "new-name" && s.Version == 2
	})).Return(nil)

	payload := &models.LoginPassword{Login: "new", Password: "pass"}
	secret, err := service.Update("secret-123", "new-name", payload, nil)

	require.NoError(t, err)
	assert.Equal(t, "new-name", secret.Name)
	assert.Equal(t, int64(2), secret.Version)
	secretRepo.AssertExpectations(t)
}

func TestService_Update_NotFound(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	secretRepo.On("GetByID", "nonexistent").Return(nil, bolt.ErrSecretNotFound)

	payload := &models.LoginPassword{Login: "test", Password: "test"}
	secret, err := service.Update("nonexistent", "name", payload, nil)

	require.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, secret)
}

func TestService_Update_RepoError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	existing := &models.Secret{
		ID:      "secret-123",
		Version: 1,
	}

	secretRepo.On("GetByID", "secret-123").Return(existing, nil)
	secretRepo.On("Update", mock.Anything).Return(assert.AnError)

	payload := &models.LoginPassword{Login: "test", Password: "test"}
	secret, err := service.Update("secret-123", "name", payload, nil)

	require.Error(t, err)
	assert.Nil(t, secret)
}

func TestService_Delete_Success(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	secretRepo.On("Delete", "secret-123").Return(nil)

	err := service.Delete("secret-123")

	require.NoError(t, err)
	secretRepo.AssertExpectations(t)
}

func TestService_Delete_NotFound(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	secretRepo.On("Delete", "nonexistent").Return(bolt.ErrSecretNotFound)

	err := service.Delete("nonexistent")

	require.ErrorIs(t, err, ErrNotFound)
}

func TestService_Delete_RepoError(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	secretRepo.On("Delete", "secret-123").Return(assert.AnError)

	err := service.Delete("secret-123")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
}

func TestService_GetRepo(t *testing.T) {
	service, secretRepo := setupSecretsService(t)

	repo := service.GetRepo()

	assert.Equal(t, secretRepo, repo)
}

func TestService_Create_AllTypes(t *testing.T) {
	testCases := []struct {
		name       string
		secretType models.SecretType
		payload    interface{}
	}{
		{
			name:       "LoginPassword",
			secretType: models.SecretTypeLoginPassword,
			payload:    &models.LoginPassword{Login: "user", Password: "pass"},
		},
		{
			name:       "TextData",
			secretType: models.SecretTypeTextData,
			payload:    &models.TextData{Text: "some text"},
		},
		{
			name:       "BinaryData",
			secretType: models.SecretTypeBinaryData,
			payload:    &models.BinaryData{FileName: "file.txt", Data: []byte("content")},
		},
		{
			name:       "BankCard",
			secretType: models.SecretTypeBankCard,
			payload:    &models.BankCard{Number: "4111111111111111", Holder: "John Doe"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, secretRepo := setupSecretsService(t)

			secretRepo.On("Create", mock.MatchedBy(func(s *models.Secret) bool {
				return s.Type == tc.secretType
			})).Return(nil)

			secret, err := service.Create(tc.name, tc.secretType, tc.payload, nil)

			require.NoError(t, err)
			assert.Equal(t, tc.secretType, secret.Type)
		})
	}
}
