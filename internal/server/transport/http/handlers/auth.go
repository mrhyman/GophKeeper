package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	authDomain "gophkeeper/internal/server/domain/auth"
	"gophkeeper/internal/server/transport/http/dto"
	"gophkeeper/internal/server/transport/http/response"
)

// AuthHandler обрабатывает запросы аутентификации.
type AuthHandler struct {
	authService *authDomain.Service
}

// NewAuthHandler создаёт новый AuthHandler.
func NewAuthHandler(authService *authDomain.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary     Регистрация
// @Description Создаёт нового пользователя и возвращает пару токенов
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body     dto.RegisterRequest true "Данные регистрации"
// @Success     201   {object} dto.TokenResponse
// @Failure     400   {object} response.ErrorBody
// @Failure     409   {object} response.ErrorBody
// @Router      /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	if req.Login == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "login and password required")
		return
	}

	tokens, err := h.authService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, authDomain.ErrUserExists) {
			response.Error(w, http.StatusConflict, "conflict", "user already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Login godoc
// @Summary     Аутентификация
// @Description Аутентифицирует пользователя и возвращает пару токенов
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body     dto.LoginRequest true "Данные входа"
// @Success     200   {object} dto.TokenResponse
// @Failure     400   {object} response.ErrorBody
// @Failure     401   {object} response.ErrorBody
// @Router      /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	if req.Login == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "login and password required")
		return
	}

	tokens, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, authDomain.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Refresh godoc
// @Summary     Обновление токенов
// @Description Обновляет пару токенов по refresh-токену
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body     dto.RefreshRequest true "Refresh-токен"
// @Success     200   {object} dto.TokenResponse
// @Failure     401   {object} response.ErrorBody
// @Router      /refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid refresh token")
		return
	}

	response.JSON(w, http.StatusOK, dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
