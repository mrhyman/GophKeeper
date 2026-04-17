package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repoMocks "gophkeeper/internal/client/repository/mocks"
	"gophkeeper/internal/client/transport/http"
	httpMocks "gophkeeper/internal/client/transport/mocks"
)

func setupAuthService(t *testing.T) (*Service, *repoMocks.AuthRepository, *httpMocks.HTTPClient) {
	authRepo := new(repoMocks.AuthRepository)
	httpClient := new(httpMocks.HTTPClient)
	service := NewService(authRepo, httpClient)
	return service, authRepo, httpClient
}

func TestService_Register_Success(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	httpClient.On("Register", "testuser", "password123").
		Return(&http.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}, nil)

	authRepo.On("SaveTokens", "access-token", "refresh-token").Return(nil)
	httpClient.On("SetAccessToken", "access-token").Return()

	err := service.Register("testuser", "password123")

	require.NoError(t, err)
	httpClient.AssertExpectations(t)
	authRepo.AssertExpectations(t)
}

func TestService_Register_HTTPError(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	httpClient.On("Register", "testuser", "password123").
		Return(nil, assert.AnError)

	err := service.Register("testuser", "password123")

	require.Error(t, err)
	authRepo.AssertNotCalled(t, "SaveTokens")
}

func TestService_Register_SaveTokensError(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	httpClient.On("Register", "testuser", "password123").
		Return(&http.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}, nil)

	authRepo.On("SaveTokens", "access-token", "refresh-token").Return(assert.AnError)

	err := service.Register("testuser", "password123")

	require.Error(t, err)
	httpClient.AssertNotCalled(t, "SetAccessToken")
}

func TestService_Login_Success(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	httpClient.On("Login", "testuser", "password123").
		Return(&http.TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		}, nil)

	authRepo.On("SaveTokens", "access-token", "refresh-token").Return(nil)
	httpClient.On("SetAccessToken", "access-token").Return()

	err := service.Login("testuser", "password123")

	require.NoError(t, err)
	httpClient.AssertExpectations(t)
	authRepo.AssertExpectations(t)
}

func TestService_Login_HTTPError(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	httpClient.On("Login", "testuser", "password123").
		Return(nil, assert.AnError)

	err := service.Login("testuser", "password123")

	require.Error(t, err)
	authRepo.AssertNotCalled(t, "SaveTokens")
}

func TestService_Logout_Success(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	httpClient.On("SetAccessToken", "").Return()
	authRepo.On("DeleteTokens").Return(nil)

	err := service.Logout()

	require.NoError(t, err)
	httpClient.AssertExpectations(t)
	authRepo.AssertExpectations(t)
}

func TestService_EnsureAuthenticated_ValidToken(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	authRepo.On("GetTokens").Return("valid-access", "refresh", nil)
	httpClient.On("SetAccessToken", "valid-access").Return()
	httpClient.On("Ping").Return(nil)

	err := service.EnsureAuthenticated()

	require.NoError(t, err)
	httpClient.AssertNotCalled(t, "Refresh")
}

func TestService_EnsureAuthenticated_RefreshNeeded(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	authRepo.On("GetTokens").Return("expired-access", "valid-refresh", nil)
	httpClient.On("SetAccessToken", "expired-access").Return().Once()
	httpClient.On("Ping").Return(assert.AnError)
	httpClient.On("Refresh", "valid-refresh").
		Return(&http.TokenPair{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
		}, nil)
	httpClient.On("SetAccessToken", "new-access").Return().Once()
	authRepo.On("SaveTokens", "new-access", "new-refresh").Return(nil)

	err := service.EnsureAuthenticated()

	require.NoError(t, err)
	httpClient.AssertExpectations(t)
	authRepo.AssertExpectations(t)
}

func TestService_EnsureAuthenticated_NoTokens(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	authRepo.On("GetTokens").Return("", "", assert.AnError)

	err := service.EnsureAuthenticated()

	require.ErrorIs(t, err, ErrNotAuthenticated)
	httpClient.AssertNotCalled(t, "SetAccessToken")
}

func TestService_EnsureAuthenticated_RefreshFailed(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	authRepo.On("GetTokens").Return("expired-access", "invalid-refresh", nil)
	httpClient.On("SetAccessToken", "expired-access").Return()
	httpClient.On("Ping").Return(assert.AnError)
	httpClient.On("Refresh", "invalid-refresh").Return(nil, assert.AnError)

	err := service.EnsureAuthenticated()

	require.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestService_EnsureAuthenticated_SaveRefreshedTokensError(t *testing.T) {
	service, authRepo, httpClient := setupAuthService(t)

	authRepo.On("GetTokens").Return("expired-access", "valid-refresh", nil)
	httpClient.On("SetAccessToken", mock.Anything).Return()
	httpClient.On("Ping").Return(assert.AnError)
	httpClient.On("Refresh", "valid-refresh").
		Return(&http.TokenPair{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
		}, nil)
	authRepo.On("SaveTokens", "new-access", "new-refresh").Return(assert.AnError)

	err := service.EnsureAuthenticated()

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotAuthenticated)
}
