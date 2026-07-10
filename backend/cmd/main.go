package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/anterajatech/warehouse-api/internal/app"
	"github.com/anterajatech/warehouse-api/internal/cache"
	"github.com/anterajatech/warehouse-api/internal/config"
	"github.com/anterajatech/warehouse-api/internal/database"
)

func main() {
	// Load configuration from environment variables / .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Root context with graceful shutdown handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Database ---
	pool, err := database.New(ctx, &cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations automatically on startup so `docker compose up` is fully self-contained.
	log.Println("running database migrations...")
	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations complete")

	// --- Cache (optional) ---
	cch := cache.New(&cfg.Redis)
	if cch.Enabled() {
		if err := cch.Ping(ctx); err != nil {
			log.Printf("WARNING: redis ping failed (caching disabled): %v", err)
		} else {
			log.Println("redis cache connected")
		}
	} else {
		log.Println("redis cache disabled (REDIS_ENABLED=false)")
	}

	// --- Application container (composition root) ---
	container := app.NewContainer(cfg, pool, cch)

	// --- HTTP server ---
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      container.Router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Run server in a goroutine so we can shut it down gracefully.
	go func() {
		log.Printf("server starting on :%s (env=%s)", cfg.Server.Port, cfg.Server.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("shutdown signal received, gracefully stopping...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}

	log.Println("server stopped cleanly")
}
