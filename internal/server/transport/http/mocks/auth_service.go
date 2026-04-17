package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	authDomain "gophkeeper/internal/server/domain/auth"
	"gophkeeper/internal/server/transport/http/types"
)

// AuthService - мок для types.AuthService.
type AuthService struct {
	mock.Mock
}

func (m *AuthService) Register(ctx context.Context, login, password string) (*authDomain.TokenPair, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.TokenPair), args.Error(1)
}

func (m *AuthService) Login(ctx context.Context, login, password string) (*authDomain.TokenPair, error) {
	args := m.Called(ctx, login, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.TokenPair), args.Error(1)
}

func (m *AuthService) Refresh(ctx context.Context, refreshToken string) (*authDomain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.TokenPair), args.Error(1)
}

func (m *AuthService) ParseAccessToken(tokenStr string) (*authDomain.Claims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authDomain.Claims), args.Error(1)
}

// Ensure AuthService implements types.AuthService.
var _ types.AuthService = (*AuthService)(nil)
