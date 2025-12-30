package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/epps11/goguard/internal/api"
	"github.com/epps11/goguard/internal/config"
	"github.com/epps11/goguard/internal/database"
	"github.com/epps11/goguard/internal/services/llm"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logging
	setupLogging(cfg.Logging)

	log.Info().
		Str("version", "1.0.0").
		Str("mode", cfg.Server.Mode).
		Msg("Starting GoGuard AI Governance Data Plane")

	// Initialize database connection
	var repo *database.Repository
	db, err := database.NewFromEnv()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to database - running without persistent settings")
	} else {
		repo = database.NewRepository(db)
		log.Info().Msg("Database connected - dashboard settings will be used")
	}

	// Initialize LLM client (optional)
	var llmClient *llm.Client
	if cfg.LLM.APIKey != "" {
		llmClient, err = llm.NewClient(cfg.LLM)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize LLM client - running without LLM forwarding")
		} else {
			log.Info().
				Str("provider", cfg.LLM.Provider).
				Str("model", cfg.LLM.Model).
				Msg("LLM client initialized")
		}
	} else {
		log.Info().Msg("No LLM API key configured - will use dashboard settings or per-request configuration")
	}

	// Create router with database repository for dynamic settings
	router := api.NewRouter(cfg, llmClient, repo)

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router.Engine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("address", addr).Msg("Server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Cleanup
	if llmClient != nil {
		llmClient.Close()
	}
	if db != nil {
		db.Close()
	}

	log.Info().Msg("Server stopped")
}

func setupLogging(cfg config.LoggingConfig) {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set output format
	if cfg.Format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	// Set output path if specified
	if cfg.OutputPath != "" {
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.Logger = log.Output(file)
		}
	}
}
