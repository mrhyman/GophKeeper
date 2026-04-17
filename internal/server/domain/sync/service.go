package sync

import (
	"context"
	"fmt"
	"time"

	"gophkeeper/internal/server/repository"
	"gophkeeper/pkg/models"
)

// SyncRequest содержит данные запроса на синхронизацию от клиента.
type SyncRequest struct {
	LastSync      int64
	ClientSecrets []*models.Secret
}

// SyncResponse содержит результат синхронизации для клиента.
type SyncResponse struct {
	ServerSecrets []*models.Secret
	SyncTimestamp int64
}

// Service реализует бизнес-логику синхронизации.
type Service struct {
	secretRepo repository.SecretRepository
}

// NewService создаёт новый sync-сервис.
func NewService(secretRepo repository.SecretRepository) *Service {
	return &Service{secretRepo: secretRepo}
}

// Sync выполняет двустороннюю синхронизацию.
// Стратегия разрешения конфликтов: last-write-wins по полю version.
// Если version совпадает — побеждает серверная версия.
func (s *Service) Sync(ctx context.Context, userID string, req *SyncRequest) (*SyncResponse, error) {
	syncTime := time.Now().Unix()

	if err := s.applyClientChanges(ctx, userID, req.ClientSecrets); err != nil {
		return nil, fmt.Errorf("apply client changes: %w", err)
	}

	serverSecrets, err := s.secretRepo.GetModifiedAfter(ctx, userID, req.LastSync)
	if err != nil {
		return nil, fmt.Errorf("get server changes: %w", err)
	}

	return &SyncResponse{
		ServerSecrets: serverSecrets,
		SyncTimestamp: syncTime,
	}, nil
}

// applyClientChanges применяет изменения от клиента на сервере.
func (s *Service) applyClientChanges(ctx context.Context, userID string, clientSecrets []*models.Secret) error {
	for _, cs := range clientSecrets {
		cs.UserID = userID

		existing, err := s.secretRepo.GetByID(ctx, userID, cs.ID)
		if err != nil {
			if err := s.secretRepo.Create(ctx, cs); err != nil {
				return fmt.Errorf("create secret %s: %w", cs.ID, err)
			}
			continue
		}

		if cs.Version > existing.Version {
			if cs.IsDeleted {
				if err := s.secretRepo.Delete(ctx, userID, cs.ID); err != nil {
					return fmt.Errorf("delete secret %s: %w", cs.ID, err)
				}
			} else {
				if err := s.secretRepo.Update(ctx, cs); err != nil {
					return fmt.Errorf("update secret %s: %w", cs.ID, err)
				}
			}
		}
	}

	return nil
}
