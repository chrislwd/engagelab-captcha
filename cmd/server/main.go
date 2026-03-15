package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/engagelab/captcha/internal/config"
	"github.com/engagelab/captcha/internal/repository"
	"github.com/engagelab/captcha/internal/router"
)

func main() {
	// Initialize logger.
	var logger *zap.Logger
	var err error

	cfg := config.Load()

	if cfg.IsProduction() {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	sugar.Infow("starting EngageLab CAPTCHA server",
		"environment", cfg.Environment,
		"port", cfg.Port,
	)

	// Attempt PostgreSQL connection (optional, fallback to in-memory).
	if cfg.DatabaseURL != "" {
		sugar.Infow("PostgreSQL URL configured", "url", maskDSN(cfg.DatabaseURL))
		// In production, you would initialize pgxpool here.
		// For now, we use in-memory storage regardless.
		sugar.Infow("using in-memory storage (PostgreSQL integration deferred)")
	} else {
		sugar.Infow("no DATABASE_URL configured, using in-memory storage")
	}

	// Attempt Redis connection (optional, fallback to in-memory).
	if cfg.RedisURL != "" {
		sugar.Infow("Redis URL configured", "url", maskDSN(cfg.RedisURL))
		sugar.Infow("using in-memory counters (Redis integration deferred)")
	} else {
		sugar.Infow("no REDIS_URL configured, using in-memory counters")
	}

	// Initialize in-memory store with seed data.
	store := repository.NewMemoryStore()
	sugar.Infow("in-memory store initialized with seed data")

	// Build router.
	r := router.New(cfg, store)

	// Create HTTP server.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine.
	go func() {
		sugar.Infow("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalw("server failed to start", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	sugar.Infow("received shutdown signal", "signal", sig.String())

	// Give outstanding requests 10 seconds to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalw("server forced to shutdown", "error", err)
	}

	sugar.Infow("server exited gracefully")
}

// maskDSN hides credentials in a connection string for logging.
func maskDSN(dsn string) string {
	if len(dsn) > 20 {
		return dsn[:10] + "****" + dsn[len(dsn)-6:]
	}
	return "****"
}
