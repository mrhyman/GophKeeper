package main

import (
	"fmt"
	"log"

	"gophkeeper/internal/client/app"
	"gophkeeper/internal/client/config"
	"gophkeeper/pkg/buildinfo"
)

func main() {
	fmt.Println(buildinfo.String())

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("client error: %v", err)
	}
}
