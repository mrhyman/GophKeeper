package mocks

import (
	"github.com/stretchr/testify/mock"

	"gophkeeper/internal/client/transport/http"
	"gophkeeper/pkg/models"
)

type HTTPClient struct {
	mock.Mock
}

func (m *HTTPClient) Register(login, password string) (*http.TokenPair, error) {
	args := m.Called(login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.TokenPair), args.Error(1)
}

func (m *HTTPClient) Login(login, password string) (*http.TokenPair, error) {
	args := m.Called(login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.TokenPair), args.Error(1)
}

func (m *HTTPClient) Refresh(refreshToken string) (*http.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.TokenPair), args.Error(1)
}

func (m *HTTPClient) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *HTTPClient) SetAccessToken(token string) {
	m.Called(token)
}

func (m *HTTPClient) Sync(lastSync int64, secrets []*models.Secret) (*http.SyncResponse, error) {
	args := m.Called(lastSync, secrets)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.SyncResponse), args.Error(1)
}
