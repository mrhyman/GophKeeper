package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/repository/mocks"
	"gophkeeper/internal/server/repository/postgres"
	"gophkeeper/pkg/models"
)

func setupAuthService(t *testing.T) (*Service, *mocks.UserRepository) {
	userRepo := new(mocks.UserRepository)

	cfg := config.JWTConfig{
		Secret:     "test-secret-key-32-bytes-long!!",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	}

	service := NewService(userRepo, cfg)

	return service, userRepo
}

func TestService_Register_Success(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = "user-123"
		}).
		Return(nil)

	tokens, err := service.Register(ctx, "testuser", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	userRepo.AssertExpectations(t)
}

func TestService_Register_UserExists(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Return(postgres.ErrUserExists)

	tokens, err := service.Register(ctx, "existinguser", "password123")

	require.ErrorIs(t, err, ErrUserExists)
	assert.Nil(t, tokens)
	userRepo.AssertExpectations(t)
}

func TestService_Register_RepoError(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Return(assert.AnError)

	tokens, err := service.Register(ctx, "testuser", "password123")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrUserExists)
	assert.Nil(t, tokens)
	userRepo.AssertExpectations(t)
}

func TestService_Login_Success(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	userRepo.On("GetByLogin", ctx, "testuser").
		Return(&models.User{
			ID:           "user-123",
			Login:        "testuser",
			PasswordHash: string(hashedPassword),
		}, nil)

	tokens, err := service.Login(ctx, "testuser", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	userRepo.AssertExpectations(t)
}

func TestService_Login_UserNotFound(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	userRepo.On("GetByLogin", ctx, "nonexistent").
		Return(nil, postgres.ErrUserNotFound)

	tokens, err := service.Login(ctx, "nonexistent", "password123")

	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, tokens)
	userRepo.AssertExpectations(t)
}

func TestService_Login_WrongPassword(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	userRepo.On("GetByLogin", ctx, "testuser").
		Return(&models.User{
			ID:           "user-123",
			Login:        "testuser",
			PasswordHash: string(hashedPassword),
		}, nil)

	tokens, err := service.Login(ctx, "testuser", "wrongpassword")

	require.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, tokens)
	userRepo.AssertExpectations(t)
}

func TestService_Login_RepoError(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	userRepo.On("GetByLogin", ctx, "testuser").
		Return(nil, assert.AnError)

	tokens, err := service.Login(ctx, "testuser", "password123")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, tokens)
	userRepo.AssertExpectations(t)
}

func TestService_Refresh_Success(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = "user-123"
		}).
		Return(nil)

	originalTokens, err := service.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	// Небольшая пауза, чтобы timestamp изменился
	time.Sleep(10 * time.Millisecond)

	// Обновляем токены
	newTokens, err := service.Refresh(ctx, originalTokens.RefreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
}

func TestService_Refresh_InvalidToken(t *testing.T) {
	service, _ := setupAuthService(t)
	ctx := context.Background()

	tokens, err := service.Refresh(ctx, "invalid-token")

	require.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, tokens)
}

func TestService_Refresh_ExpiredToken(t *testing.T) {
	userRepo := new(mocks.UserRepository)

	// Создаём сервис с очень коротким TTL
	cfg := config.JWTConfig{
		Secret:     "test-secret-key-32-bytes-long!!",
		AccessTTL:  1 * time.Millisecond,
		RefreshTTL: 1 * time.Millisecond,
	}

	service := NewService(userRepo, cfg)
	ctx := context.Background()

	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = "user-123"
		}).
		Return(nil)

	originalTokens, err := service.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	// Ждём истечения токена
	time.Sleep(10 * time.Millisecond)

	tokens, err := service.Refresh(ctx, originalTokens.RefreshToken)

	require.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, tokens)
}

func TestService_ParseAccessToken_Success(t *testing.T) {
	service, userRepo := setupAuthService(t)
	ctx := context.Background()

	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = "user-123"
		}).
		Return(nil)

	tokens, err := service.Register(ctx, "testuser", "password123")
	require.NoError(t, err)

	claims, err := service.ParseAccessToken(tokens.AccessToken)

	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
}

func TestService_ParseAccessToken_Invalid(t *testing.T) {
	service, _ := setupAuthService(t)

	claims, err := service.ParseAccessToken("invalid-token")

	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_ParseAccessToken_WrongSignature(t *testing.T) {
	service, _ := setupAuthService(t)

	// Токен подписан другим секретом
	otherService := NewService(nil, config.JWTConfig{
		Secret:     "different-secret-key-32-bytes!!",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	})

	// Генерируем токен вручную через другой сервис
	userRepo := new(mocks.UserRepository)
	otherService.userRepo = userRepo

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = "user-123"
		}).
		Return(nil)

	tokens, _ := otherService.Register(context.Background(), "test", "pass")

	claims, err := service.ParseAccessToken(tokens.AccessToken)

	require.Error(t, err)
	assert.Nil(t, claims)
}
