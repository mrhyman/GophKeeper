package http

import "gophkeeper/pkg/models"

// HTTPClient описывает контракт HTTP-клиента.
type HTTPClient interface {
	Register(login, password string) (*TokenPair, error)
	Login(login, password string) (*TokenPair, error)
	Refresh(refreshToken string) (*TokenPair, error)
	Ping() error
	SetAccessToken(token string)
	Sync(lastSync int64, secrets []*models.Secret) (*SyncResponse, error)
}
