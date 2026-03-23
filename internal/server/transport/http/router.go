package http

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"gophkeeper/internal/server/config"
	authDomain "gophkeeper/internal/server/domain/auth"
	secretsDomain "gophkeeper/internal/server/domain/secrets"
	syncDomain "gophkeeper/internal/server/domain/sync"
	"gophkeeper/internal/server/transport/http/handlers"
	"gophkeeper/internal/server/transport/http/middleware"
)

// NewRouter создаёт и конфигурирует chi-роутер.
func NewRouter(
	authService *authDomain.Service,
	secretsService *secretsDomain.Service,
	syncService *syncDomain.Service,
	jwtCfg config.JWTConfig,
	logger *zap.Logger,
) chi.Router {
	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	secretsHandler := handlers.NewSecretsHandler(secretsService)
	syncHandler := handlers.NewSyncHandler(syncService)

	// Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		// Публичные
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)

		// Защищённые
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))

			r.Route("/secrets", func(r chi.Router) {
				r.Post("/", secretsHandler.Create)
				r.Get("/", secretsHandler.List)
				r.Get("/{id}", secretsHandler.Get)
				r.Put("/{id}", secretsHandler.Update)
				r.Delete("/{id}", secretsHandler.Delete)
			})

			r.Post("/sync", syncHandler.Sync)
		})
	})

	return r
}
