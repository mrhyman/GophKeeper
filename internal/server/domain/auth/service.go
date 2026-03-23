package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/repository"
	"gophkeeper/internal/server/repository/postgres"
	"gophkeeper/pkg/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

// TokenPair содержит пару access и refresh токенов.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Claims описывает payload JWT-токена.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// Service реализует бизнес-логику аутентификации.
type Service struct {
	userRepo repository.UserRepository
	jwtCfg   config.JWTConfig
}

// NewService создаёт новый auth-сервис.
func NewService(userRepo repository.UserRepository, jwtCfg config.JWTConfig) *Service {
	return &Service{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

// Register регистрирует нового пользователя и возвращает пару токенов.
func (s *Service) Register(ctx context.Context, login, password string) (*TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Login:        login,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, postgres.ErrUserExists) {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateTokenPair(user.ID)
}

// Login аутентифицирует пользователя и возвращает пару токенов.
func (s *Service) Login(ctx context.Context, login, password string) (*TokenPair, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user.ID)
}

// Refresh обновляет пару токенов по валидному refresh-токену.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return s.generateTokenPair(claims.UserID)
}

// ParseAccessToken валидирует access-токен и возвращает claims.
// Используется в middleware авторизации.
func (s *Service) ParseAccessToken(tokenStr string) (*Claims, error) {
	return s.parseToken(tokenStr)
}

// generateTokenPair создаёт пару JWT-токенов.
func (s *Service) generateTokenPair(userID string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtCfg.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: userID,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtCfg.RefreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: userID,
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// parseToken парсит и валидирует JWT-токен.
func (s *Service) parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
