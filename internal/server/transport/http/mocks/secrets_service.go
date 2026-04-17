package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gophkeeper/internal/server/transport/http/types"
	"gophkeeper/pkg/models"
)

// SecretsService - мок для types.SecretsService.
type SecretsService struct {
	mock.Mock
}

func (m *SecretsService) Create(ctx context.Context, secret *models.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *SecretsService) Update(ctx context.Context, secret *models.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *SecretsService) Delete(ctx context.Context, userID, secretID string) error {
	args := m.Called(ctx, userID, secretID)
	return args.Error(0)
}

func (m *SecretsService) Get(ctx context.Context, userID, secretID string) (*models.Secret, error) {
	args := m.Called(ctx, userID, secretID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Secret), args.Error(1)
}

func (m *SecretsService) List(ctx context.Context, userID string) ([]*models.Secret, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Secret), args.Error(1)
}

// Ensure SecretsService implements types.SecretsService.
var _ types.SecretsService = (*SecretsService)(nil)
