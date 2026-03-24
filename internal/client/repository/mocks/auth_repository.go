package mocks

import (
	"github.com/stretchr/testify/mock"
)

type AuthRepository struct {
	mock.Mock
}

func (m *AuthRepository) SaveTokens(access, refresh string) error {
	args := m.Called(access, refresh)
	return args.Error(0)
}

func (m *AuthRepository) GetTokens() (string, string, error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

func (m *AuthRepository) DeleteTokens() error {
	args := m.Called()
	return args.Error(0)
}

func (m *AuthRepository) SaveLastSync(timestamp int64) error {
	args := m.Called(timestamp)
	return args.Error(0)
}

func (m *AuthRepository) GetLastSync() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}
