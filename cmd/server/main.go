package main

import (
	"log"

	"gophkeeper/internal/server/app"
	"gophkeeper/internal/server/config"
)

// @title       GophKeeper API
// @version     1.0
// @description Менеджер паролей — серверная часть
// @host        localhost:8080
// @BasePath    /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
