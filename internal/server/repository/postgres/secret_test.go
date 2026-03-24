package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophkeeper/pkg/models"
)

func setupSecretRepo(t *testing.T) (pgxmock.PgxPoolIface, *SecretRepository) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	db := &DB{Pool: mock}
	repo := NewSecretRepository(db)

	return mock, repo
}

func TestSecretRepository_Create_Success(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	secret := &models.Secret{
		UserID:   "user-123",
		Name:     "my-secret",
		Type:     models.SecretTypeLoginPassword,
		Data:     []byte("encrypted-data"),
		Metadata: map[string]string{"key": "value"},
	}

	rows := pgxmock.NewRows([]string{"id"}).AddRow("secret-456")

	mock.ExpectQuery(`INSERT INTO secrets`).
		WithArgs(
			secret.UserID,
			secret.Name,
			int(secret.Type),
			secret.Data,
			pgxmock.AnyArg(), // metadata JSON
			pgxmock.AnyArg(), // updated_at
		).
		WillReturnRows(rows)

	err := repo.Create(context.Background(), secret)

	require.NoError(t, err)
	assert.Equal(t, "secret-456", secret.ID)
	assert.Equal(t, int64(1), secret.Version)
	assert.NotZero(t, secret.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_Create_DBError(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	secret := &models.Secret{
		UserID:   "user-123",
		Name:     "my-secret",
		Type:     models.SecretTypeLoginPassword,
		Data:     []byte("encrypted-data"),
		Metadata: map[string]string{},
	}

	mock.ExpectQuery(`INSERT INTO secrets`).
		WithArgs(
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnError(assert.AnError)

	err := repo.Create(context.Background(), secret)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_GetByID_Success(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	metadata := map[string]string{"key": "value"}
	metaJSON, _ := json.Marshal(metadata)

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "name", "type", "data", "metadata", "version", "updated_at", "is_deleted",
	}).AddRow(
		"secret-456",
		"user-123",
		"my-secret",
		int(models.SecretTypeLoginPassword),
		[]byte("encrypted-data"),
		metaJSON,
		int64(1),
		time.Now().Unix(),
		false,
	)

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets`).
		WithArgs("secret-456", "user-123").
		WillReturnRows(rows)

	secret, err := repo.GetByID(context.Background(), "user-123", "secret-456")

	require.NoError(t, err)
	assert.Equal(t, "secret-456", secret.ID)
	assert.Equal(t, "user-123", secret.UserID)
	assert.Equal(t, "my-secret", secret.Name)
	assert.Equal(t, models.SecretTypeLoginPassword, secret.Type)
	assert.Equal(t, "value", secret.Metadata["key"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_GetByID_NotFound(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets`).
		WithArgs("nonexistent", "user-123").
		WillReturnError(pgx.ErrNoRows)

	secret, err := repo.GetByID(context.Background(), "user-123", "nonexistent")

	require.ErrorIs(t, err, ErrSecretNotFound)
	assert.Nil(t, secret)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_List_Success(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	metadata := map[string]string{}
	metaJSON, _ := json.Marshal(metadata)
	now := time.Now().Unix()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "name", "type", "data", "metadata", "version", "updated_at", "is_deleted",
	}).
		AddRow("secret-1", "user-123", "secret1", int(models.SecretTypeLoginPassword), []byte("data1"), metaJSON, int64(1), now, false).
		AddRow("secret-2", "user-123", "secret2", int(models.SecretTypeTextData), []byte("data2"), metaJSON, int64(2), now, false)

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets WHERE user_id = \$1 AND is_deleted = FALSE`).
		WithArgs("user-123").
		WillReturnRows(rows)

	secrets, err := repo.List(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.Equal(t, "secret-1", secrets[0].ID)
	assert.Equal(t, "secret-2", secrets[1].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_List_Empty(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "name", "type", "data", "metadata", "version", "updated_at", "is_deleted",
	})

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets WHERE user_id = \$1 AND is_deleted = FALSE`).
		WithArgs("user-123").
		WillReturnRows(rows)

	secrets, err := repo.List(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Empty(t, secrets)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_List_DBError(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets WHERE user_id = \$1 AND is_deleted = FALSE`).
		WithArgs("user-123").
		WillReturnError(assert.AnError)

	secrets, err := repo.List(context.Background(), "user-123")

	require.Error(t, err)
	assert.Nil(t, secrets)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_Update_Success(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	secret := &models.Secret{
		ID:       "secret-456",
		UserID:   "user-123",
		Name:     "updated-secret",
		Type:     models.SecretTypeTextData,
		Data:     []byte("new-encrypted-data"),
		Metadata: map[string]string{"updated": "true"},
	}

	newVersion := int64(2)
	newUpdatedAt := time.Now().Unix()

	rows := pgxmock.NewRows([]string{"version", "updated_at"}).
		AddRow(newVersion, newUpdatedAt)

	mock.ExpectQuery(`UPDATE secrets`).
		WithArgs(
			secret.Name,
			int(secret.Type),
			secret.Data,
			pgxmock.AnyArg(), // metadata JSON
			pgxmock.AnyArg(), // updated_at
			secret.ID,
			secret.UserID,
		).
		WillReturnRows(rows)

	err := repo.Update(context.Background(), secret)

	require.NoError(t, err)
	assert.Equal(t, newVersion, secret.Version)
	assert.Equal(t, newUpdatedAt, secret.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_Update_NotFound(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	secret := &models.Secret{
		ID:       "nonexistent",
		UserID:   "user-123",
		Name:     "secret",
		Type:     models.SecretTypeTextData,
		Data:     []byte("data"),
		Metadata: map[string]string{},
	}

	mock.ExpectQuery(`UPDATE secrets`).
		WithArgs(
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnError(pgx.ErrNoRows)

	err := repo.Update(context.Background(), secret)

	require.ErrorIs(t, err, ErrSecretNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_Delete_Success(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	mock.ExpectExec(`UPDATE secrets SET is_deleted = TRUE`).
		WithArgs(pgxmock.AnyArg(), "secret-456", "user-123").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.Delete(context.Background(), "user-123", "secret-456")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_Delete_NotFound(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	mock.ExpectExec(`UPDATE secrets SET is_deleted = TRUE`).
		WithArgs(pgxmock.AnyArg(), "nonexistent", "user-123").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.Delete(context.Background(), "user-123", "nonexistent")

	require.ErrorIs(t, err, ErrSecretNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_Delete_DBError(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	mock.ExpectExec(`UPDATE secrets SET is_deleted = TRUE`).
		WithArgs(pgxmock.AnyArg(), "secret-456", "user-123").
		WillReturnError(assert.AnError)

	err := repo.Delete(context.Background(), "user-123", "secret-456")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrSecretNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_GetModifiedAfter_Success(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	metadata := map[string]string{}
	metaJSON, _ := json.Marshal(metadata)
	since := time.Now().Add(-1 * time.Hour).Unix()
	now := time.Now().Unix()

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "name", "type", "data", "metadata", "version", "updated_at", "is_deleted",
	}).
		AddRow("secret-1", "user-123", "secret1", int(models.SecretTypeLoginPassword), []byte("data1"), metaJSON, int64(1), now, false).
		AddRow("secret-2", "user-123", "secret2", int(models.SecretTypeTextData), []byte("data2"), metaJSON, int64(2), now, true) // deleted

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets WHERE user_id = \$1 AND updated_at > \$2`).
		WithArgs("user-123", since).
		WillReturnRows(rows)

	secrets, err := repo.GetModifiedAfter(context.Background(), "user-123", since)

	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.False(t, secrets[0].IsDeleted)
	assert.True(t, secrets[1].IsDeleted)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_GetModifiedAfter_DBError(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets WHERE user_id = \$1 AND updated_at > \$2`).
		WithArgs("user-123", int64(0)).
		WillReturnError(assert.AnError)

	secrets, err := repo.GetModifiedAfter(context.Background(), "user-123", 0)

	require.Error(t, err)
	assert.Nil(t, secrets)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_GetByID_InvalidMetadata(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	// Невалидный JSON в metadata
	invalidJSON := []byte(`{invalid}`)

	rows := pgxmock.NewRows([]string{
		"id", "user_id", "name", "type", "data", "metadata", "version", "updated_at", "is_deleted",
	}).AddRow(
		"secret-456",
		"user-123",
		"my-secret",
		int(models.SecretTypeLoginPassword),
		[]byte("encrypted-data"),
		invalidJSON,
		int64(1),
		time.Now().Unix(),
		false,
	)

	mock.ExpectQuery(`SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted FROM secrets`).
		WithArgs("secret-456", "user-123").
		WillReturnRows(rows)

	secret, err := repo.GetByID(context.Background(), "user-123", "secret-456")

	require.Error(t, err)
	assert.Nil(t, secret)
	assert.Contains(t, err.Error(), "unmarshal metadata")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepository_Create_NilMetadata(t *testing.T) {
	mock, repo := setupSecretRepo(t)
	defer mock.Close()

	secret := &models.Secret{
		UserID:   "user-123",
		Name:     "my-secret",
		Type:     models.SecretTypeLoginPassword,
		Data:     []byte("encrypted-data"),
		Metadata: nil, // nil metadata
	}

	rows := pgxmock.NewRows([]string{"id"}).AddRow("secret-456")

	mock.ExpectQuery(`INSERT INTO secrets`).
		WithArgs(
			secret.UserID,
			secret.Name,
			int(secret.Type),
			secret.Data,
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnRows(rows)

	err := repo.Create(context.Background(), secret)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
