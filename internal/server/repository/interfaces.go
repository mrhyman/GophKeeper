package repository

import (
	"context"

	"gophkeeper/pkg/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}

type SecretRepository interface {
	Create(ctx context.Context, secret *models.Secret) error
	GetByID(ctx context.Context, userID, secretID string) (*models.Secret, error)
	List(ctx context.Context, userID string) ([]*models.Secret, error)
	Update(ctx context.Context, secret *models.Secret) error
	Delete(ctx context.Context, userID, secretID string) error
	GetModifiedAfter(ctx context.Context, userID string, since int64) ([]*models.Secret, error)
}
