package main

import (
	"fmt"
	"log"

	"github.com/yourusername/virallens/backend/internal/config"
	"github.com/yourusername/virallens/backend/internal/wire"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize server via Wire DI
	r, err := wire.InitializeServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting Virallens Backend Server on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
