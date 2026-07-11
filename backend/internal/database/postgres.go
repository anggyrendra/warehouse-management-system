package database

import (
	"context"
	"fmt"
	"time"

	"github.com/anterajatech/warehouse-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// New creates and returns a configured PostgreSQL connection pool.
// The pool is pinged once during setup so connection errors surface early.
func New(ctx context.Context, cfg *config.DBConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	// Sensible pool defaults for a small service.
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	// Verify connectivity before handing the pool back.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}
