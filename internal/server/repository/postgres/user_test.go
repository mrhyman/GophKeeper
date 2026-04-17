package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophkeeper/pkg/models"
)

func setupUserRepo(t *testing.T) (pgxmock.PgxPoolIface, *UserRepository) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	db := &DB{Pool: mock}
	repo := NewUserRepository(db)

	return mock, repo
}

func TestUserRepository_Create_Success(t *testing.T) {
	mock, repo := setupUserRepo(t)
	defer mock.Close()

	user := &models.User{
		Login:        "testuser",
		PasswordHash: "hashedpassword",
	}

	rows := pgxmock.NewRows([]string{"id", "created_at"}).
		AddRow("user-123", time.Now().Unix())

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Login, user.PasswordHash).
		WillReturnRows(rows)

	err := repo.Create(context.Background(), user)

	require.NoError(t, err)
	assert.Equal(t, "user-123", user.ID)
	assert.NotZero(t, user.CreatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_AlreadyExists(t *testing.T) {
	mock, repo := setupUserRepo(t)
	defer mock.Close()

	user := &models.User{
		Login:        "existinguser",
		PasswordHash: "hashedpassword",
	}

	// Код 23505 = unique_violation
	pgErr := &pgconn.PgError{Code: "23505"}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Login, user.PasswordHash).
		WillReturnError(pgErr)

	err := repo.Create(context.Background(), user)

	require.ErrorIs(t, err, ErrUserExists)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DBError(t *testing.T) {
	mock, repo := setupUserRepo(t)
	defer mock.Close()

	user := &models.User{
		Login:        "testuser",
		PasswordHash: "hashedpassword",
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Login, user.PasswordHash).
		WillReturnError(assert.AnError)

	err := repo.Create(context.Background(), user)

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrUserExists)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByLogin_Success(t *testing.T) {
	mock, repo := setupUserRepo(t)
	defer mock.Close()

	expectedUser := &models.User{
		ID:           "user-123",
		Login:        "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now().Unix(),
	}

	rows := pgxmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
		AddRow(expectedUser.ID, expectedUser.Login, expectedUser.PasswordHash, expectedUser.CreatedAt)

	mock.ExpectQuery(`SELECT id, login, password_hash, created_at FROM users`).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := repo.GetByLogin(context.Background(), "testuser")

	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Login, user.Login)
	assert.Equal(t, expectedUser.PasswordHash, user.PasswordHash)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByLogin_NotFound(t *testing.T) {
	mock, repo := setupUserRepo(t)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, login, password_hash, created_at FROM users`).
		WithArgs("nonexistent").
		WillReturnError(pgx.ErrNoRows)

	user, err := repo.GetByLogin(context.Background(), "nonexistent")

	require.ErrorIs(t, err, ErrUserNotFound)
	assert.Nil(t, user)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByLogin_DBError(t *testing.T) {
	mock, repo := setupUserRepo(t)
	defer mock.Close()

	mock.ExpectQuery(`SELECT id, login, password_hash, created_at FROM users`).
		WithArgs("testuser").
		WillReturnError(assert.AnError)

	user, err := repo.GetByLogin(context.Background(), "testuser")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrUserNotFound)
	assert.Nil(t, user)
	require.NoError(t, mock.ExpectationsWereMet())
}
