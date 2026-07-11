package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/anterajatech/warehouse-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LocationRepository handles location persistence.
type LocationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{db: db}
}

// List returns all warehouse locations, ordered by code.
func (r *LocationRepository) List(ctx context.Context) ([]model.Location, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, zone, type, created_at
		FROM locations
		ORDER BY code
	`)
	if err != nil {
		return nil, fmt.Errorf("location list: %w", err)
	}
	defer rows.Close()

	var locs []model.Location
	for rows.Next() {
		var l model.Location
		if err := rows.Scan(&l.ID, &l.Code, &l.Zone, &l.Type, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("location list scan: %w", err)
		}
		locs = append(locs, l)
	}
	return locs, rows.Err()
}

// GetByID fetches a single location by its primary key.
func (r *LocationRepository) GetByID(ctx context.Context, id int64) (*model.Location, error) {
	const q = `SELECT id, code, zone, type, created_at FROM locations WHERE id = $1`
	var l model.Location
	err := r.db.QueryRow(ctx, q, id).
		Scan(&l.ID, &l.Code, &l.Zone, &l.Type, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("location get by id: %w", err)
	}
	return &l, nil
}
