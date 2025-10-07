package main

import (
	"log"
)

func main() {
	// Init logger
	InitLogger()

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Config load failed: %v", err)
	}

	// Start tray GUI in goroutine
	go StartTray(cfg)

	// Start HTTP server
	StartServer(cfg)
}