package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	secretsDomain "gophkeeper/internal/server/domain/secrets"
	"gophkeeper/internal/server/transport/http/dto"
	"gophkeeper/internal/server/transport/http/middleware"
	"gophkeeper/internal/server/transport/http/response"
	"gophkeeper/internal/server/transport/http/types"
	"gophkeeper/pkg/models"
)

// SecretsHandler обрабатывает запросы к секретам.
type SecretsHandler struct {
	secretsService types.SecretsService
}

// NewSecretsHandler создаёт новый SecretsHandler.
func NewSecretsHandler(secretsService types.SecretsService) *SecretsHandler {
	return &SecretsHandler{secretsService: secretsService}
}

// Create godoc
// @Summary     Создать секрет
// @Tags        secrets
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body     dto.CreateSecretRequest true "Данные секрета"
// @Success     201   {object} dto.SecretResponse
// @Failure     400   {object} response.ErrorBody
// @Router      /secrets [post]
func (h *SecretsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req dto.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	secret := &models.Secret{
		UserID:   userID,
		Name:     req.Name,
		Type:     models.SecretType(req.Type),
		Data:     req.Data,
		Metadata: req.Metadata,
	}

	if err := h.secretsService.Create(r.Context(), secret); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, toSecretResponse(secret))
}

// List godoc
// @Summary     Список секретов
// @Tags        secrets
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} dto.SecretListResponse
// @Router      /secrets [get]
func (h *SecretsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	secrets, err := h.secretsService.List(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	resp := dto.SecretListResponse{
		Secrets: make([]dto.SecretResponse, 0, len(secrets)),
	}
	for _, s := range secrets {
		resp.Secrets = append(resp.Secrets, toSecretResponse(s))
	}

	response.JSON(w, http.StatusOK, resp)
}

// Get godoc
// @Summary     Получить секрет
// @Tags        secrets
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID секрета"
// @Success     200 {object} dto.SecretResponse
// @Failure     404 {object} response.ErrorBody
// @Router      /secrets/{id} [get]
func (h *SecretsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	secretID := chi.URLParam(r, "id")

	secret, err := h.secretsService.Get(r.Context(), userID, secretID)
	if err != nil {
		if errors.Is(err, secretsDomain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "secret not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toSecretResponse(secret))
}

// Update godoc
// @Summary     Обновить секрет
// @Tags        secrets
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string                true "ID секрета"
// @Param       input body dto.UpdateSecretRequest true "Новые данные"
// @Success     200   {object} dto.SecretResponse
// @Failure     404   {object} response.ErrorBody
// @Router      /secrets/{id} [put]
func (h *SecretsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	secretID := chi.URLParam(r, "id")

	var req dto.UpdateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	secret := &models.Secret{
		ID:       secretID,
		UserID:   userID,
		Name:     req.Name,
		Type:     models.SecretType(req.Type),
		Data:     req.Data,
		Metadata: req.Metadata,
	}

	if err := h.secretsService.Update(r.Context(), secret); err != nil {
		if errors.Is(err, secretsDomain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "secret not found")
			return
		}
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, toSecretResponse(secret))
}

// Delete godoc
// @Summary     Удалить секрет
// @Tags        secrets
// @Security    BearerAuth
// @Param       id path string true "ID секрета"
// @Success     204
// @Failure     404 {object} response.ErrorBody
// @Router      /secrets/{id} [delete]
func (h *SecretsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	secretID := chi.URLParam(r, "id")

	if err := h.secretsService.Delete(r.Context(), userID, secretID); err != nil {
		if errors.Is(err, secretsDomain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "secret not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toSecretResponse конвертирует доменную модель в DTO.
func toSecretResponse(s *models.Secret) dto.SecretResponse {
	return dto.SecretResponse{
		ID:        s.ID,
		Name:      s.Name,
		Type:      int(s.Type),
		Data:      s.Data,
		Metadata:  s.Metadata,
		Version:   s.Version,
		UpdatedAt: s.UpdatedAt,
		IsDeleted: s.IsDeleted,
	}
}
