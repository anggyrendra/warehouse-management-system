package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/anterajatech/warehouse-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors that the service layer maps to HTTP status codes.
var (
	ErrNotFound      = errors.New("repository: record not found")
	ErrDuplicateSKU  = errors.New("repository: duplicate sku")
	ErrForeignKey    = errors.New("repository: foreign key violation")
)

// ItemRepository handles all item persistence.
type ItemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

// Create inserts a new item and returns the row with generated fields.
func (r *ItemRepository) Create(ctx context.Context, item *model.Item) error {
	const q = `
		INSERT INTO items (sku, name, category, unit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, q, item.SKU, item.Name, item.Category, item.Unit).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateSKU
		}
		return fmt.Errorf("item create: %w", err)
	}
	return nil
}

// GetByID fetches a single non-deleted item by its primary key.
func (r *ItemRepository) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	const q = `
		SELECT id, sku, name, category, unit, created_at, updated_at, deleted_at
		FROM items
		WHERE id = $1 AND deleted_at IS NULL
	`
	var item model.Item
	err := r.db.QueryRow(ctx, q, id).
		Scan(&item.ID, &item.SKU, &item.Name, &item.Category, &item.Unit,
			&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("item get by id: %w", err)
	}
	return &item, nil
}

// GetBySKU fetches a single non-deleted item by its SKU — used to enforce
// uniqueness at the service layer without relying solely on the DB constraint.
func (r *ItemRepository) GetBySKU(ctx context.Context, sku string) (*model.Item, error) {
	const q = `
		SELECT id, sku, name, category, unit, created_at, updated_at, deleted_at
		FROM items
		WHERE sku = $1 AND deleted_at IS NULL
	`
	var item model.Item
	err := r.db.QueryRow(ctx, q, sku).
		Scan(&item.ID, &item.SKU, &item.Name, &item.Category, &item.Unit,
			&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("item get by sku: %w", err)
	}
	return &item, nil
}

// List returns a page of non-deleted items, optionally filtered by category.
// Results are ordered by created_at descending so newest items appear first.
func (r *ItemRepository) List(ctx context.Context, category string, limit, offset int) (*model.PaginatedItems, error) {
	// Fetch the page of rows
	rows, err := r.db.Query(ctx, `
		SELECT id, sku, name, category, unit, created_at, updated_at, deleted_at
		FROM items
		WHERE deleted_at IS NULL
		  AND ($1 = '' OR category = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("item list query: %w", err)
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var it model.Item
		if err := rows.Scan(&it.ID, &it.SKU, &it.Name, &it.Category, &it.Unit,
			&it.CreatedAt, &it.UpdatedAt, &it.DeletedAt); err != nil {
			return nil, fmt.Errorf("item list scan: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("item list rows: %w", err)
	}

	// Fetch total count (with the same filter) for pagination metadata
	var total int64
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM items
		WHERE deleted_at IS NULL
		  AND ($1 = '' OR category = $1)
	`, category).Scan(&total); err != nil {
		return nil, fmt.Errorf("item list count: %w", err)
	}

	if items == nil {
		items = []model.Item{}
	}

	return &model.PaginatedItems{Items: items, Total: total}, nil
}

// Update mutates an existing non-deleted item's editable fields.
func (r *ItemRepository) Update(ctx context.Context, item *model.Item) error {
	const q = `
		UPDATE items
		SET sku = $1, name = $2, category = $3, unit = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING created_at, updated_at, deleted_at
	`
	err := r.db.QueryRow(ctx, q, item.SKU, item.Name, item.Category, item.Unit, item.ID).
		Scan(&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if isUniqueViolation(err) {
			return ErrDuplicateSKU
		}
		return fmt.Errorf("item update: %w", err)
	}
	return nil
}

// SoftDelete sets deleted_at instead of removing the row, preserving referential integrity.
func (r *ItemRepository) SoftDelete(ctx context.Context, id int64) error {
	const q = `UPDATE items SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("item soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListCategories returns distinct non-deleted item categories for the frontend filter dropdown.
func (r *ItemRepository) ListCategories(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT category FROM items
		WHERE deleted_at IS NULL
		ORDER BY category
	`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("list categories scan: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

// isUniqueViolation inspects a pgx error to see whether it is a unique-constraint violation,
// which we use to detect duplicate SKUs.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}
