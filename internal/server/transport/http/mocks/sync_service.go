package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	syncDomain "gophkeeper/internal/server/domain/sync"
	"gophkeeper/internal/server/transport/http/types"
)

// SyncService - мок для types.SyncService.
type SyncService struct {
	mock.Mock
}

func (m *SyncService) Sync(ctx context.Context, userID string, req *syncDomain.SyncRequest) (*syncDomain.SyncResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*syncDomain.SyncResponse), args.Error(1)
}

// Ensure SyncService implements types.SyncService.
var _ types.SyncService = (*SyncService)(nil)
