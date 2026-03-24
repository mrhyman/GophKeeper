package bolt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestNew_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	// Проверяем, что файл создан
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestNew_InvalidPath(t *testing.T) {
	// Пытаемся создать БД в несуществующей директории без прав
	db, err := New("/nonexistent/path/test.db")

	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestDB_Close(t *testing.T) {
	db := setupTestDB(t)

	err := db.Close()
	assert.NoError(t, err)
}
