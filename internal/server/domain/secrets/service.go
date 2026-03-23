package secrets

import (
	"context"
	"errors"
	"fmt"

	"gophkeeper/internal/server/repository"
	"gophkeeper/internal/server/repository/postgres"
	"gophkeeper/pkg/models"
)

var (
	ErrNotFound = errors.New("secret not found")
)

// Service реализует бизнес-логику работы с секретами.
type Service struct {
	secretRepo repository.SecretRepository
}

// NewService создаёт новый secrets-сервис.
func NewService(secretRepo repository.SecretRepository) *Service {
	return &Service{secretRepo: secretRepo}
}

// Create создаёт новый секрет для пользователя.
func (s *Service) Create(ctx context.Context, secret *models.Secret) error {
	if err := s.validate(secret); err != nil {
		return err
	}

	if err := s.secretRepo.Create(ctx, secret); err != nil {
		return fmt.Errorf("create secret: %w", err)
	}

	return nil
}

// Get возвращает секрет по ID с проверкой владельца.
func (s *Service) Get(ctx context.Context, userID, secretID string) (*models.Secret, error) {
	secret, err := s.secretRepo.GetByID(ctx, userID, secretID)
	if err != nil {
		if errors.Is(err, postgres.ErrSecretNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}

	return secret, nil
}

// List возвращает все секреты пользователя.
func (s *Service) List(ctx context.Context, userID string) ([]*models.Secret, error) {
	secrets, err := s.secretRepo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	return secrets, nil
}

// Update обновляет секрет.
func (s *Service) Update(ctx context.Context, secret *models.Secret) error {
	if err := s.validate(secret); err != nil {
		return err
	}

	if err := s.secretRepo.Update(ctx, secret); err != nil {
		if errors.Is(err, postgres.ErrSecretNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("update secret: %w", err)
	}

	return nil
}

// Delete удаляет секрет (soft delete).
func (s *Service) Delete(ctx context.Context, userID, secretID string) error {
	if err := s.secretRepo.Delete(ctx, userID, secretID); err != nil {
		if errors.Is(err, postgres.ErrSecretNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete secret: %w", err)
	}

	return nil
}

// validate выполняет базовую валидацию секрета.
func (s *Service) validate(secret *models.Secret) error {
	if secret.Name == "" {
		return errors.New("secret name is required")
	}

	if secret.Type < models.SecretTypeLoginPassword || secret.Type > models.SecretTypeBankCard {
		return errors.New("invalid secret type")
	}

	if len(secret.Data) == 0 {
		return errors.New("secret data is required")
	}

	return nil
}
