package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gophkeeper/pkg/models"
)

type SecretRepository struct {
	mock.Mock
}

func (m *SecretRepository) Create(ctx context.Context, secret *models.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *SecretRepository) GetByID(ctx context.Context, userID, secretID string) (*models.Secret, error) {
	args := m.Called(ctx, userID, secretID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Secret), args.Error(1)
}

func (m *SecretRepository) List(ctx context.Context, userID string) ([]*models.Secret, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Secret), args.Error(1)
}

func (m *SecretRepository) Update(ctx context.Context, secret *models.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *SecretRepository) Delete(ctx context.Context, userID, secretID string) error {
	args := m.Called(ctx, userID, secretID)
	return args.Error(0)
}

func (m *SecretRepository) GetModifiedAfter(ctx context.Context, userID string, since int64) ([]*models.Secret, error) {
	args := m.Called(ctx, userID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Secret), args.Error(1)
}
