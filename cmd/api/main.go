package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MayaCris/stock-info-app/internal/infrastructure/config"
	"github.com/MayaCris/stock-info-app/internal/infrastructure/database/cockroachdb"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	log.Printf("🚀 Starting %s in %s mode on port %s", 
		cfg.App.Name, cfg.App.Env, cfg.App.Port)

	// Initialize database connection
	db, err := cockroachdb.NewConnection(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("❌ Error closing database: %v", err)
		}
	}()

	// Test database health
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if !db.IsHealthy(ctx) {
		log.Fatal("❌ Database health check failed")
	}

	// Print database stats
	if stats, err := db.GetStats(); err == nil {
		log.Printf("📊 Database stats: %+v", stats)
	}

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✅ Application started successfully")
	log.Println("Press Ctrl+C to shutdown...")

	// Wait for shutdown signal
	<-quit
	log.Println("🔄 Shutting down gracefully...")

	// Here you would add cleanup for your HTTP server, workers, etc.
	
	log.Println("👋 Application stopped")
}