package auth

import (
	"errors"
	"fmt"

	"gophkeeper/internal/client/repository"
	"gophkeeper/internal/client/transport/http"
)

var (
	ErrNotAuthenticated = errors.New("not authenticated")
)

type Service struct {
	authRepo   repository.AuthRepository
	httpClient *http.Client
}

func NewService(authRepo repository.AuthRepository, httpClient *http.Client) *Service {
	return &Service{
		authRepo:   authRepo,
		httpClient: httpClient,
	}
}

func (s *Service) Register(login, password string) error {
	tokens, err := s.httpClient.Register(login, password)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	// Сохраняем токены локально
	if err := s.authRepo.SaveTokens(tokens.AccessToken, tokens.RefreshToken); err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}

	// Устанавливаем токен в HTTP-клиент для последующих запросов
	s.httpClient.SetAccessToken(tokens.AccessToken)

	return nil
}

func (s *Service) Login(login, password string) error {
	tokens, err := s.httpClient.Login(login, password)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	if err := s.authRepo.SaveTokens(tokens.AccessToken, tokens.RefreshToken); err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}

	// Устанавливаем токен в HTTP-клиент
	s.httpClient.SetAccessToken(tokens.AccessToken)

	return nil
}

func (s *Service) Logout() error {
	// Очищаем токен из HTTP-клиента
	s.httpClient.SetAccessToken("")
	return s.authRepo.DeleteTokens()
}

// EnsureAuthenticated проверяет наличие токенов и при необходимости обновляет.
func (s *Service) EnsureAuthenticated() error {
	access, refresh, err := s.authRepo.GetTokens()
	if err != nil {
		return ErrNotAuthenticated
	}

	// Пробуем текущий access-токен
	s.httpClient.SetAccessToken(access)
	if s.httpClient.Ping() == nil {
		return nil
	}

	// Access протух — пробуем refresh
	tokens, err := s.httpClient.Refresh(refresh)
	if err != nil {
		return ErrNotAuthenticated
	}

	s.httpClient.SetAccessToken(tokens.AccessToken)

	if err := s.authRepo.SaveTokens(tokens.AccessToken, tokens.RefreshToken); err != nil {
		return fmt.Errorf("save refreshed tokens: %w", err)
	}

	return nil
}
