package repository

import (
	"context"
	"fmt"

	"github.com/anterajatech/warehouse-api/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StockRepository handles stock persistence.
// Stock mutations are performed inside a single SQL statement (upsert) to keep
// the operation atomic at the database level.
type StockRepository struct {
	db *pgxpool.Pool
}

func NewStockRepository(db *pgxpool.Pool) *StockRepository {
	return &StockRepository{db: db}
}

// Receive performs an upsert: if a stock row exists for the (item_id, location_id)
// pair, the quantity is incremented; otherwise a new row is created.
// The non-negative constraint is enforced at the service layer AND via a DB check.
func (r *StockRepository) Receive(ctx context.Context, req model.StockReceive) (*model.Stock, error) {
	const q = `
		INSERT INTO stock (item_id, location_id, qty, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (item_id, location_id)
		DO UPDATE
			SET qty = stock.qty + EXCLUDED.qty,
			    updated_at = NOW()
		RETURNING id, item_id, location_id, qty, updated_at
	`
	var s model.Stock
	err := r.db.QueryRow(ctx, q, req.ItemID, req.LocationID, req.Qty).
		Scan(&s.ID, &s.ItemID, &s.LocationID, &s.Qty, &s.UpdatedAt)
	if err != nil {
		// A check-constraint violation (qty < 0) surfaces here if the service
		// layer validation was somehow bypassed — defense in depth.
		return nil, fmt.Errorf("stock receive: %w", err)
	}
	return &s, nil
}

// GetByItem returns all stock rows for a given item, joined with location details.
// Only non-deleted items are considered.
func (r *StockRepository) GetByItem(ctx context.Context, itemID int64) ([]model.StockWithDetail, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.item_id, i.sku, i.name, s.location_id, l.code, l.zone, s.qty, s.updated_at
		FROM stock s
		JOIN items i     ON i.id = s.item_id
		JOIN locations l ON l.id = s.location_id
		WHERE s.item_id = $1 AND i.deleted_at IS NULL
		ORDER BY l.code
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("stock get by item: %w", err)
	}
	defer rows.Close()

	var stocks []model.StockWithDetail
	for rows.Next() {
		var s model.StockWithDetail
		if err := rows.Scan(&s.ID, &s.ItemID, &s.ItemSKU, &s.ItemName,
			&s.LocationID, &s.LocationCode, &s.Zone, &s.Qty, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("stock get by item scan: %w", err)
		}
		stocks = append(stocks, s)
	}

	if stocks == nil {
		stocks = []model.StockWithDetail{}
	}
	return stocks, rows.Err()
}

// ItemExists is a lightweight existence check used to validate stock-receive
// requests without fetching the full item row.
func (r *StockRepository) ItemExists(ctx context.Context, itemID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM items WHERE id = $1 AND deleted_at IS NULL)
	`, itemID).Scan(&exists)
	if err != nil {
		// ignore pgx.ErrNoRows — EXISTS always returns a row
		return false, fmt.Errorf("stock item exists: %w", err)
	}
	return exists, nil
}

// LocationExists checks whether a location exists by id.
func (r *StockRepository) LocationExists(ctx context.Context, locationID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1)
	`, locationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("stock location exists: %w", err)
	}
	return exists, nil
}

