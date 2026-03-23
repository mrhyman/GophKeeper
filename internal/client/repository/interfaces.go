package repository

import (
	"gophkeeper/pkg/models"
)

// SecretRepository описывает контракт локального хранилища секретов.
type SecretRepository interface {
	Create(secret *models.Secret) error
	GetByID(secretID string) (*models.Secret, error)
	List() ([]*models.Secret, error)
	Update(secret *models.Secret) error
	Delete(secretID string) error
	UpsertBatch(secrets []*models.Secret) error
}

// AuthRepository хранит данные сессии (токены, timestamp синхронизации).
type AuthRepository interface {
	SaveTokens(access, refresh string) error
	GetTokens() (access, refresh string, err error)
	DeleteTokens() error
	SaveLastSync(timestamp int64) error
	GetLastSync() (int64, error)
}
