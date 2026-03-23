package middleware

import (
	"context"
	"net/http"
	"strings"

	authDomain "gophkeeper/internal/server/domain/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// Auth проверяет JWT-токен и кладёт userID в контекст.
func Auth(authService *authDomain.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if tokenStr == header {
				http.Error(w, `{"error":"invalid auth header"}`, http.StatusUnauthorized)
				return
			}

			claims, err := authService.ParseAccessToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID извлекает userID из контекста.
func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}
