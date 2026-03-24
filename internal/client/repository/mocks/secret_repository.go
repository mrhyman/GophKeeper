package mocks

import (
	"github.com/stretchr/testify/mock"

	"gophkeeper/pkg/models"
)

type SecretRepository struct {
	mock.Mock
}

func (m *SecretRepository) Create(secret *models.Secret) error {
	args := m.Called(secret)
	return args.Error(0)
}

func (m *SecretRepository) GetByID(secretID string) (*models.Secret, error) {
	args := m.Called(secretID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Secret), args.Error(1)
}

func (m *SecretRepository) List() ([]*models.Secret, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Secret), args.Error(1)
}

func (m *SecretRepository) Update(secret *models.Secret) error {
	args := m.Called(secret)
	return args.Error(0)
}

func (m *SecretRepository) Delete(secretID string) error {
	args := m.Called(secretID)
	return args.Error(0)
}

func (m *SecretRepository) UpsertBatch(secrets []*models.Secret) error {
	args := m.Called(secrets)
	return args.Error(0)
}
