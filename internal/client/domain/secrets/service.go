package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gophkeeper/internal/client/repository"
	boltRepo "gophkeeper/internal/client/repository/bolt"
	"gophkeeper/pkg/crypto"
	"gophkeeper/pkg/models"
)

var (
	ErrNotFound = errors.New("secret not found")
)

// Service реализует клиентскую логику работы с секретами.
type Service struct {
	secretRepo     repository.SecretRepository
	masterPassword string
}

// NewService создаёт новый клиентский secrets-сервис.
func NewService(secretRepo repository.SecretRepository, masterPassword string) *Service {
	return &Service{
		secretRepo:     secretRepo,
		masterPassword: masterPassword,
	}
}

// Create создаёт новый секрет: сериализует типизированные данные, шифрует и сохраняет в локальное хранилище.
func (s *Service) Create(name string, secretType models.SecretType, payload interface{}, metadata map[string]string) (*models.Secret, error) {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	encrypted, err := crypto.Encrypt(plaintext, s.masterPassword)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	secret := &models.Secret{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      secretType,
		Data:      encrypted,
		Metadata:  metadata,
		Version:   1,
		UpdatedAt: time.Now().Unix(),
	}

	if err := s.secretRepo.Create(secret); err != nil {
		return nil, fmt.Errorf("save secret: %w", err)
	}

	return secret, nil
}

// Get возвращает секрет по ID.
func (s *Service) Get(secretID string) (*models.Secret, error) {
	secret, err := s.secretRepo.GetByID(secretID)
	if err != nil {
		if errors.Is(err, boltRepo.ErrSecretNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}

	return secret, nil
}

// Decrypt расшифровывает данные секрета в указанную структуру.
func (s *Service) Decrypt(secret *models.Secret, dest interface{}) error {
	plaintext, err := crypto.Decrypt(secret.Data, s.masterPassword)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	if err := json.Unmarshal(plaintext, dest); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}

// List возвращает все секреты.
func (s *Service) List() ([]*models.Secret, error) {
	secrets, err := s.secretRepo.List()
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	return secrets, nil
}

// Update обновляет секрет: шифрует новые данные, инкрементирует версию.
func (s *Service) Update(secretID, name string, payload interface{}, metadata map[string]string) (*models.Secret, error) {
	existing, err := s.secretRepo.GetByID(secretID)
	if err != nil {
		if errors.Is(err, boltRepo.ErrSecretNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	encrypted, err := crypto.Encrypt(plaintext, s.masterPassword)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	existing.Name = name
	existing.Data = encrypted
	existing.Metadata = metadata
	existing.Version++
	existing.UpdatedAt = time.Now().Unix()

	if err := s.secretRepo.Update(existing); err != nil {
		return nil, fmt.Errorf("update secret: %w", err)
	}

	return existing, nil
}

// Delete удаляет секрет (soft delete).
func (s *Service) Delete(secretID string) error {
	if err := s.secretRepo.Delete(secretID); err != nil {
		if errors.Is(err, boltRepo.ErrSecretNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete secret: %w", err)
	}

	return nil
}

// GetRepo возвращает репозиторий секретов.
// Используется при пересоздании сервиса с новым мастер-паролем.
func (s *Service) GetRepo() repository.SecretRepository {
	return s.secretRepo
}
