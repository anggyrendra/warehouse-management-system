package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations reads all .sql files from the migrations directory and applies
// them in lexical order. Each file runs in its own transaction.
//
// The migrations directory defaults to "./migrations" (relative to the working
// directory) but can be overridden via the MIGRATIONS_DIR environment variable.
// This keeps the SQL files at the repository root as required by the deliverables
// while still being embeddable in a Docker image.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "./migrations"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}

	// Collect and sort .sql files by filename so 001_… runs before 002_…
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		if err := applyMigration(ctx, pool, filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", filepath.Base(path), err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for %s: %w", filepath.Base(path), err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck — rollback is a no-op if commit ran

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("apply migration %s: %w", filepath.Base(path), err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", filepath.Base(path), err)
	}

	log.Printf("[MIGRATION] applied %s", filepath.Base(path))
	return nil
}
