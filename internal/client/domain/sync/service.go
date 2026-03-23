package sync

import (
	"fmt"

	"gophkeeper/internal/client/repository"
	httpTransport "gophkeeper/internal/client/transport/http"
	"gophkeeper/pkg/models"
)

// Service реализует клиентскую логику синхронизации.
type Service struct {
	secretRepo repository.SecretRepository
	authRepo   repository.AuthRepository
	httpClient *httpTransport.Client
}

// SyncResult содержит статистику синхронизации.
type SyncResult struct {
	Uploaded   int
	Downloaded int
}

// NewService создаёт новый клиентский sync-сервис.
func NewService(
	secretRepo repository.SecretRepository,
	authRepo repository.AuthRepository,
	httpClient *httpTransport.Client,
) *Service {
	return &Service{
		secretRepo: secretRepo,
		authRepo:   authRepo,
		httpClient: httpClient,
	}
}

// Sync выполняет двустороннюю синхронизацию с сервером.
func (s *Service) Sync() (*SyncResult, error) {
	lastSync, err := s.authRepo.GetLastSync()
	if err != nil {
		return nil, fmt.Errorf("get last sync: %w", err)
	}

	localSecrets, err := s.getLocalChanges(lastSync)
	if err != nil {
		return nil, fmt.Errorf("get local changes: %w", err)
	}

	resp, err := s.httpClient.Sync(lastSync, localSecrets)
	if err != nil {
		return nil, fmt.Errorf("sync with server: %w", err)
	}

	if len(resp.ServerSecrets) > 0 {
		if err := s.secretRepo.UpsertBatch(resp.ServerSecrets); err != nil {
			return nil, fmt.Errorf("apply server changes: %w", err)
		}
	}

	if err := s.authRepo.SaveLastSync(resp.SyncTimestamp); err != nil {
		return nil, fmt.Errorf("save sync timestamp: %w", err)
	}

	return &SyncResult{
		Uploaded:   len(localSecrets),
		Downloaded: len(resp.ServerSecrets),
	}, nil
}

// getLocalChanges возвращает секреты, изменённые после lastSync.
func (s *Service) getLocalChanges(lastSync int64) ([]*models.Secret, error) {
	all, err := s.secretRepo.List()
	if err != nil {
		return nil, err
	}

	var changed []*models.Secret
	for _, secret := range all {
		if secret.UpdatedAt > lastSync {
			changed = append(changed, secret)
		}
	}

	return changed, nil
}
