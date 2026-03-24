package types

import (
	"context"

	authDomain "gophkeeper/internal/server/domain/auth"
	syncDomain "gophkeeper/internal/server/domain/sync"
	"gophkeeper/pkg/models"
)

// AuthService определяет интерфейс для сервиса аутентификации.
type AuthService interface {
	Register(ctx context.Context, login, password string) (*authDomain.TokenPair, error)
	Login(ctx context.Context, login, password string) (*authDomain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*authDomain.TokenPair, error)
	ParseAccessToken(tokenStr string) (*authDomain.Claims, error)
}

// SecretsService определяет интерфейс для сервиса секретов.
type SecretsService interface {
	Create(ctx context.Context, secret *models.Secret) error
	Update(ctx context.Context, secret *models.Secret) error
	Delete(ctx context.Context, userID, secretID string) error
	Get(ctx context.Context, userID, secretID string) (*models.Secret, error)
	List(ctx context.Context, userID string) ([]*models.Secret, error)
}

// SyncService определяет интерфейс для сервиса синхронизации.
type SyncService interface {
	Sync(ctx context.Context, userID string, req *syncDomain.SyncRequest) (*syncDomain.SyncResponse, error)
}
