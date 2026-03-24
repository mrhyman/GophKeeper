package bolt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthRepo(t *testing.T) *AuthRepository {
	db := setupTestDB(t)
	return NewAuthRepository(db)
}

func TestAuthRepository_SaveTokens_Success(t *testing.T) {
	repo := setupAuthRepo(t)

	err := repo.SaveTokens("access-token-123", "refresh-token-456")
	require.NoError(t, err)

	access, refresh, err := repo.GetTokens()
	require.NoError(t, err)
	assert.Equal(t, "access-token-123", access)
	assert.Equal(t, "refresh-token-456", refresh)
}

func TestAuthRepository_GetTokens_NoTokens(t *testing.T) {
	repo := setupAuthRepo(t)

	access, refresh, err := repo.GetTokens()

	require.ErrorIs(t, err, ErrNoTokens)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
}

func TestAuthRepository_SaveTokens_Overwrite(t *testing.T) {
	repo := setupAuthRepo(t)

	// Сохраняем первые токены
	require.NoError(t, repo.SaveTokens("old-access", "old-refresh"))

	// Перезаписываем
	require.NoError(t, repo.SaveTokens("new-access", "new-refresh"))

	access, refresh, err := repo.GetTokens()
	require.NoError(t, err)
	assert.Equal(t, "new-access", access)
	assert.Equal(t, "new-refresh", refresh)
}

func TestAuthRepository_DeleteTokens_Success(t *testing.T) {
	repo := setupAuthRepo(t)

	// Сохраняем токены
	require.NoError(t, repo.SaveTokens("access", "refresh"))

	// Удаляем
	err := repo.DeleteTokens()
	require.NoError(t, err)

	// Проверяем, что токенов нет
	_, _, err = repo.GetTokens()
	require.ErrorIs(t, err, ErrNoTokens)
}

func TestAuthRepository_DeleteTokens_NoTokens(t *testing.T) {
	repo := setupAuthRepo(t)

	// Удаление без токенов не должно быть ошибкой
	err := repo.DeleteTokens()
	require.NoError(t, err)
}

func TestAuthRepository_SaveLastSync_Success(t *testing.T) {
	repo := setupAuthRepo(t)

	err := repo.SaveLastSync(1234567890)
	require.NoError(t, err)

	ts, err := repo.GetLastSync()
	require.NoError(t, err)
	assert.Equal(t, int64(1234567890), ts)
}

func TestAuthRepository_GetLastSync_NoSync(t *testing.T) {
	repo := setupAuthRepo(t)

	ts, err := repo.GetLastSync()

	require.NoError(t, err)
	assert.Equal(t, int64(0), ts) // 0 означает, что синхронизации не было
}

func TestAuthRepository_SaveLastSync_Overwrite(t *testing.T) {
	repo := setupAuthRepo(t)

	require.NoError(t, repo.SaveLastSync(1000))
	require.NoError(t, repo.SaveLastSync(2000))

	ts, err := repo.GetLastSync()
	require.NoError(t, err)
	assert.Equal(t, int64(2000), ts)
}

func TestAuthRepository_TokensAndSyncIndependent(t *testing.T) {
	repo := setupAuthRepo(t)

	// Сохраняем и токены, и sync
	require.NoError(t, repo.SaveTokens("access", "refresh"))
	require.NoError(t, repo.SaveLastSync(12345))

	// Удаляем токены
	require.NoError(t, repo.DeleteTokens())

	// Sync должен остаться
	ts, err := repo.GetLastSync()
	require.NoError(t, err)
	assert.Equal(t, int64(12345), ts)

	// Токены удалены
	_, _, err = repo.GetTokens()
	require.ErrorIs(t, err, ErrNoTokens)
}

func TestAuthRepository_LargeTimestamp(t *testing.T) {
	repo := setupAuthRepo(t)

	// Проверяем большие значения timestamp (год 2100+)
	largeTS := int64(4102444800) // 2100-01-01

	err := repo.SaveLastSync(largeTS)
	require.NoError(t, err)

	ts, err := repo.GetLastSync()
	require.NoError(t, err)
	assert.Equal(t, largeTS, ts)
}

func TestAuthRepository_EmptyTokens(t *testing.T) {
	repo := setupAuthRepo(t)

	// Пустые строки сохраняются, но GetTokens вернёт ошибку
	err := repo.SaveTokens("", "")
	require.NoError(t, err)

	_, _, err = repo.GetTokens()
	require.ErrorIs(t, err, ErrNoTokens)
}

func TestAuthRepository_LongTokens(t *testing.T) {
	repo := setupAuthRepo(t)

	// JWT токены могут быть длинными
	longAccess := make([]byte, 2048)
	longRefresh := make([]byte, 2048)
	for i := range longAccess {
		longAccess[i] = 'a'
		longRefresh[i] = 'b'
	}

	err := repo.SaveTokens(string(longAccess), string(longRefresh))
	require.NoError(t, err)

	access, refresh, err := repo.GetTokens()
	require.NoError(t, err)
	assert.Len(t, access, 2048)
	assert.Len(t, refresh, 2048)
}
