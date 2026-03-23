package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/domain/auth"
	"gophkeeper/internal/server/domain/secrets"
	domainSync "gophkeeper/internal/server/domain/sync"
	"gophkeeper/internal/server/repository/postgres"
	httpTransport "gophkeeper/internal/server/transport/http"
)

// Run инициализирует зависимости и запускает HTTP-сервер.
func Run(cfg *config.Config) error {
	// Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Repository
	db, err := postgres.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	secretRepo := postgres.NewSecretRepository(db)

	// Domain services
	authService := auth.NewService(userRepo, cfg.JWT)
	secretsService := secrets.NewService(secretRepo)
	syncService := domainSync.NewService(secretRepo)

	// HTTP
	router := httpTransport.NewRouter(authService, secretsService, syncService, cfg.JWT, logger)

	srvCfg := httpTransport.ServerConfig{
		Address:  cfg.Server.Address,
		CertFile: cfg.Server.TLS.Cert,
		KeyFile:  cfg.Server.TLS.Key,
	}
	srv := httpTransport.NewServer(srvCfg, router)

	// Graceful shutdown
	errCh := make(chan error, 1)

	go func() {
		errCh <- httpTransport.ListenAndServe(srv, cfg.Server.TLS.Cert, cfg.Server.TLS.Key)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
