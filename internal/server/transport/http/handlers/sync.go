package handlers

import (
	"encoding/json"
	"net/http"

	syncDomain "gophkeeper/internal/server/domain/sync"
	"gophkeeper/internal/server/transport/http/dto"
	"gophkeeper/internal/server/transport/http/middleware"
	"gophkeeper/internal/server/transport/http/response"
	"gophkeeper/internal/server/transport/http/types"
	"gophkeeper/pkg/models"
)

// SyncHandler обрабатывает запросы синхронизации.
type SyncHandler struct {
	syncService types.SyncService
}

// NewSyncHandler создаёт новый SyncHandler.
func NewSyncHandler(syncService types.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

// Sync godoc
// @Summary     Синхронизация
// @Description Двусторонняя синхронизация секретов
// @Tags        sync
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body     dto.SyncRequest true "Данные синхронизации"
// @Success     200   {object} dto.SyncResponse
// @Router      /sync [post]
func (h *SyncHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req dto.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	// DTO → domain
	clientSecrets := make([]*models.Secret, 0, len(req.ClientSecrets))
	for _, cs := range req.ClientSecrets {
		clientSecrets = append(clientSecrets, &models.Secret{
			ID:        cs.ID,
			UserID:    userID,
			Name:      cs.Name,
			Type:      models.SecretType(cs.Type),
			Data:      cs.Data,
			Metadata:  cs.Metadata,
			Version:   cs.Version,
			UpdatedAt: cs.UpdatedAt,
			IsDeleted: cs.IsDeleted,
		})
	}

	domainReq := &syncDomain.SyncRequest{
		LastSync:      req.LastSync,
		ClientSecrets: clientSecrets,
	}

	resp, err := h.syncService.Sync(r.Context(), userID, domainReq)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	// domain → DTO
	serverSecrets := make([]dto.SecretResponse, 0, len(resp.ServerSecrets))
	for _, s := range resp.ServerSecrets {
		serverSecrets = append(serverSecrets, toSecretResponse(s))
	}

	response.JSON(w, http.StatusOK, dto.SyncResponse{
		ServerSecrets: serverSecrets,
		SyncTimestamp: resp.SyncTimestamp,
	})
}
