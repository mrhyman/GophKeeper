package bolt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophkeeper/pkg/models"
)

func setupSecretRepo(t *testing.T) *SecretRepository {
	db := setupTestDB(t)
	return NewSecretRepository(db)
}

func TestSecretRepository_Create_Success(t *testing.T) {
	repo := setupSecretRepo(t)

	secret := &models.Secret{
		ID:       "secret-123",
		UserID:   "user-456",
		Name:     "my-secret",
		Type:     models.SecretTypeLoginPassword,
		Data:     []byte("encrypted-data"),
		Metadata: map[string]string{"key": "value"},
		Version:  1,
	}

	err := repo.Create(secret)
	require.NoError(t, err)

	// Проверяем, что секрет сохранён
	got, err := repo.GetByID("secret-123")
	require.NoError(t, err)
	assert.Equal(t, secret.ID, got.ID)
	assert.Equal(t, secret.Name, got.Name)
	assert.Equal(t, secret.Type, got.Type)
	assert.Equal(t, secret.Data, got.Data)
	assert.Equal(t, secret.Metadata, got.Metadata)
}

func TestSecretRepository_GetByID_Success(t *testing.T) {
	repo := setupSecretRepo(t)

	secret := &models.Secret{
		ID:      "secret-123",
		Name:    "test-secret",
		Type:    models.SecretTypeTextData,
		Data:    []byte("some data"),
		Version: 1,
	}
	require.NoError(t, repo.Create(secret))

	got, err := repo.GetByID("secret-123")

	require.NoError(t, err)
	assert.Equal(t, "secret-123", got.ID)
	assert.Equal(t, "test-secret", got.Name)
}

func TestSecretRepository_GetByID_NotFound(t *testing.T) {
	repo := setupSecretRepo(t)

	got, err := repo.GetByID("nonexistent")

	require.ErrorIs(t, err, ErrSecretNotFound)
	assert.Nil(t, got)
}

func TestSecretRepository_List_Success(t *testing.T) {
	repo := setupSecretRepo(t)

	secrets := []*models.Secret{
		{ID: "secret-1", Name: "first", Type: models.SecretTypeLoginPassword, Version: 1},
		{ID: "secret-2", Name: "second", Type: models.SecretTypeTextData, Version: 1},
		{ID: "secret-3", Name: "third", Type: models.SecretTypeBankCard, Version: 1},
	}

	for _, s := range secrets {
		require.NoError(t, repo.Create(s))
	}

	got, err := repo.List()

	require.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestSecretRepository_List_ExcludesDeleted(t *testing.T) {
	repo := setupSecretRepo(t)

	secrets := []*models.Secret{
		{ID: "secret-1", Name: "active", IsDeleted: false, Version: 1},
		{ID: "secret-2", Name: "deleted", IsDeleted: true, Version: 1},
	}

	for _, s := range secrets {
		require.NoError(t, repo.Create(s))
	}

	got, err := repo.List()

	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "active", got[0].Name)
}

func TestSecretRepository_List_Empty(t *testing.T) {
	repo := setupSecretRepo(t)

	got, err := repo.List()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestSecretRepository_Update_Success(t *testing.T) {
	repo := setupSecretRepo(t)

	secret := &models.Secret{
		ID:      "secret-123",
		Name:    "original",
		Type:    models.SecretTypeLoginPassword,
		Data:    []byte("original-data"),
		Version: 1,
	}
	require.NoError(t, repo.Create(secret))

	// Обновляем
	secret.Name = "updated"
	secret.Data = []byte("new-data")
	secret.Version = 2

	err := repo.Update(secret)
	require.NoError(t, err)

	// Проверяем
	got, err := repo.GetByID("secret-123")
	require.NoError(t, err)
	assert.Equal(t, "updated", got.Name)
	assert.Equal(t, []byte("new-data"), got.Data)
	assert.Equal(t, int64(2), got.Version)
}

func TestSecretRepository_Update_NotFound(t *testing.T) {
	repo := setupSecretRepo(t)

	secret := &models.Secret{
		ID:   "nonexistent",
		Name: "test",
	}

	err := repo.Update(secret)

	require.ErrorIs(t, err, ErrSecretNotFound)
}

func TestSecretRepository_Delete_Success(t *testing.T) {
	repo := setupSecretRepo(t)

	secret := &models.Secret{
		ID:      "secret-123",
		Name:    "to-delete",
		Version: 1,
	}
	require.NoError(t, repo.Create(secret))

	err := repo.Delete("secret-123")
	require.NoError(t, err)

	// Секрет всё ещё существует, но помечен удалённым
	got, err := repo.GetByID("secret-123")
	require.NoError(t, err)
	assert.True(t, got.IsDeleted)
	assert.Equal(t, int64(2), got.Version) // версия увеличилась

	// В списке не отображается
	list, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestSecretRepository_Delete_NotFound(t *testing.T) {
	repo := setupSecretRepo(t)

	err := repo.Delete("nonexistent")

	require.ErrorIs(t, err, ErrSecretNotFound)
}

func TestSecretRepository_UpsertBatch_CreateNew(t *testing.T) {
	repo := setupSecretRepo(t)

	secrets := []*models.Secret{
		{ID: "secret-1", Name: "first", Version: 1},
		{ID: "secret-2", Name: "second", Version: 1},
	}

	err := repo.UpsertBatch(secrets)
	require.NoError(t, err)

	got, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestSecretRepository_UpsertBatch_UpdateExisting(t *testing.T) {
	repo := setupSecretRepo(t)

	// Создаём секрет
	original := &models.Secret{
		ID:      "secret-1",
		Name:    "original",
		Version: 1,
	}
	require.NoError(t, repo.Create(original))

	// Upsert с новой версией
	updated := []*models.Secret{
		{ID: "secret-1", Name: "updated", Version: 2},
	}

	err := repo.UpsertBatch(updated)
	require.NoError(t, err)

	got, err := repo.GetByID("secret-1")
	require.NoError(t, err)
	assert.Equal(t, "updated", got.Name)
	assert.Equal(t, int64(2), got.Version)
}

func TestSecretRepository_UpsertBatch_Empty(t *testing.T) {
	repo := setupSecretRepo(t)

	err := repo.UpsertBatch([]*models.Secret{})

	require.NoError(t, err)
}

func TestSecretRepository_UpsertBatch_WithDeleted(t *testing.T) {
	repo := setupSecretRepo(t)

	secrets := []*models.Secret{
		{ID: "secret-1", Name: "active", IsDeleted: false, Version: 1},
		{ID: "secret-2", Name: "deleted", IsDeleted: true, Version: 1},
	}

	err := repo.UpsertBatch(secrets)
	require.NoError(t, err)

	// Оба секрета сохранены
	got1, err := repo.GetByID("secret-1")
	require.NoError(t, err)
	assert.False(t, got1.IsDeleted)

	got2, err := repo.GetByID("secret-2")
	require.NoError(t, err)
	assert.True(t, got2.IsDeleted)

	// Но в списке только активный
	list, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestSecretRepository_AllTypes(t *testing.T) {
	repo := setupSecretRepo(t)

	testCases := []struct {
		name       string
		secretType models.SecretType
	}{
		{"LoginPassword", models.SecretTypeLoginPassword},
		{"TextData", models.SecretTypeTextData},
		{"BinaryData", models.SecretTypeBinaryData},
		{"BankCard", models.SecretTypeBankCard},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			secret := &models.Secret{
				ID:      "secret-" + tc.name,
				Name:    tc.name,
				Type:    tc.secretType,
				Data:    []byte("test-data"),
				Version: 1,
			}

			err := repo.Create(secret)
			require.NoError(t, err)

			got, err := repo.GetByID(secret.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.secretType, got.Type)
		})
	}
}

func TestSecretRepository_WithMetadata(t *testing.T) {
	repo := setupSecretRepo(t)

	secret := &models.Secret{
		ID:   "secret-123",
		Name: "with-metadata",
		Metadata: map[string]string{
			"url":      "https://example.com",
			"category": "work",
			"notes":    "important",
		},
		Version: 1,
	}

	err := repo.Create(secret)
	require.NoError(t, err)

	got, err := repo.GetByID("secret-123")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", got.Metadata["url"])
	assert.Equal(t, "work", got.Metadata["category"])
	assert.Equal(t, "important", got.Metadata["notes"])
}
