package app

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"gophkeeper/internal/client/config"
	authDomain "gophkeeper/internal/client/domain/auth"
	secretsDomain "gophkeeper/internal/client/domain/secrets"
	syncDomain "gophkeeper/internal/client/domain/sync"
	"gophkeeper/internal/client/repository/bolt"
	httpTransport "gophkeeper/internal/client/transport/http"
	"gophkeeper/internal/client/tui"
)

// Run инициализирует зависимости и запускает TUI.
func Run(cfg *config.Config) error {
	// Создаём директорию хранилища
	storageDir := filepath.Dir(cfg.Storage.Path)
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}

	// Repository
	db, err := bolt.New(cfg.Storage.Path)
	if err != nil {
		return fmt.Errorf("open storage: %w", err)
	}
	defer db.Close()

	authRepo := bolt.NewAuthRepository(db)
	secretRepo := bolt.NewSecretRepository(db)

	// HTTP client с TLS
	httpClient, err := httpTransport.NewClient(cfg.Server.Address, cfg.Server.CAFile)
	if err != nil {
		return fmt.Errorf("create http client: %w", err)
	}

	// Domain services
	authService := authDomain.NewService(authRepo, httpClient)
	secretsService := secretsDomain.NewService(secretRepo, "")
	syncService := syncDomain.NewService(secretRepo, authRepo, httpClient)

	// TUI
	services := tui.Services{
		Auth:    authService,
		Secrets: secretsService,
		Sync:    syncService,
	}

	p := tea.NewProgram(
		tui.New(services),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	return nil
}
