package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/anterajatech/warehouse-api/internal/model"
	"github.com/anterajatech/warehouse-api/internal/repository"
)

// Service-level sentinel errors. The handler layer maps these to HTTP status codes.
var (
	ErrItemNotFound      = errors.New("item not found")
	ErrSKUDuplicate      = errors.New("sku already exists")
	ErrStockNotFound     = errors.New("stock not found")
	ErrInvalidStockQty   = errors.New("stock quantity must be a positive integer")
	ErrReferencedItem    = errors.New("item not found")
	ErrReferencedLocation = errors.New("location not found")
)

// ItemService implements business rules for warehouse items.
type ItemService struct {
	repo  *repository.ItemRepository
	cache *cacheWrapper
}

func NewItemService(repo *repository.ItemRepository, cw *cacheWrapper) *ItemService {
	return &ItemService{repo: repo, cache: cw}
}

// Create adds a new item. SKU uniqueness is enforced by the DB constraint
// and mapped to a service-level ErrSKUDuplicate.
func (s *ItemService) Create(ctx context.Context, item *model.Item) (*model.Item, error) {
	if err := s.repo.Create(ctx, item); err != nil {
		if errors.Is(err, repository.ErrDuplicateSKU) {
			return nil, ErrSKUDuplicate
		}
		return nil, fmt.Errorf("item service create: %w", err)
	}
	// Invalidate the list cache since a new item was added.
	s.cache.InvalidateItemList(ctx)
	return item, nil
}

// GetByID retrieves an item, using the cache when enabled.
func (s *ItemService) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	// Try cache first
	var cached model.Item
	if s.cache.enabled {
		if err := s.cache.GetItem(ctx, fmt.Sprintf("item:%d", id), &cached); err == nil {
			return &cached, nil
		}
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("item service get by id: %w", err)
	}

	// Cache the result
	s.cache.SetItem(ctx, fmt.Sprintf("item:%d", id), item)
	return item, nil
}

// List returns a page of items filtered by category.
func (s *ItemService) List(ctx context.Context, category string, page, limit int) ([]model.Item, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	result, err := s.repo.List(ctx, strings.TrimSpace(category), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("item service list: %w", err)
	}
	return result.Items, result.Total, nil
}

// Update modifies an existing item.
func (s *ItemService) Update(ctx context.Context, item *model.Item) (*model.Item, error) {
	if err := s.repo.Update(ctx, item); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrItemNotFound
		}
		if errors.Is(err, repository.ErrDuplicateSKU) {
			return nil, ErrSKUDuplicate
		}
		return nil, fmt.Errorf("item service update: %w", err)
	}
	s.cache.InvalidateItem(ctx, item.ID)
	return item, nil
}

// SoftDelete removes an item from active results without deleting the row.
func (s *ItemService) SoftDelete(ctx context.Context, id int64) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrItemNotFound
		}
		return fmt.Errorf("item service soft delete: %w", err)
	}
	s.cache.InvalidateItem(ctx, id)
	return nil
}

// ListCategories returns distinct categories for the frontend filter.
func (s *ItemService) ListCategories(ctx context.Context) ([]string, error) {
	return s.repo.ListCategories(ctx)
}

// StockService implements business rules for stock operations.
type StockService struct {
	repo      *repository.StockRepository
	itemRepo  *repository.ItemRepository
	locRepo   *repository.LocationRepository
	cache     *cacheWrapper
}

func NewStockService(
	repo *repository.StockRepository,
	itemRepo *repository.ItemRepository,
	locRepo *repository.LocationRepository,
	cw *cacheWrapper,
) *StockService {
	return &StockService{repo: repo, itemRepo: itemRepo, locRepo: locRepo, cache: cw}
}

// Receive handles inbound stock to a location.
// Business rules:
//   - qty must be a positive integer (validated here, not just at DB level)
//   - item and location must exist
//   - every mutation is logged to the application log
func (s *StockService) Receive(ctx context.Context, req model.StockReceive) (*model.Stock, error) {
	// Rule: quantity must be positive
	if req.Qty <= 0 {
		return nil, ErrInvalidStockQty
	}

	// Validate item existence
	exists, err := s.repo.ItemExists(ctx, req.ItemID)
	if err != nil {
		return nil, fmt.Errorf("stock service receive: %w", err)
	}
	if !exists {
		return nil, ErrReferencedItem
	}

	// Validate location existence
	locExists, err := s.repo.LocationExists(ctx, req.LocationID)
	if err != nil {
		return nil, fmt.Errorf("stock service receive: %w", err)
	}
	if !locExists {
		return nil, ErrReferencedLocation
	}

	stock, err := s.repo.Receive(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("stock service receive: %w", err)
	}

	// Log the mutation as required by the business rules (console/file).
	log.Printf("[STOCK-LOG] RECEIVE item_id=%d location_id=%d qty=%+d new_qty=%d",
		req.ItemID, req.LocationID, req.Qty, stock.Qty)

	// Invalidate stock cache for this item
	s.cache.Delete(ctx, fmt.Sprintf("stock:%d", req.ItemID))

	return stock, nil
}

// GetByItem returns the stock for an item across all locations.
func (s *StockService) GetByItem(ctx context.Context, itemID int64) ([]model.StockWithDetail, error) {
	// Verify the item exists first to give a clean 404.
	if _, err := s.itemRepo.GetByID(ctx, itemID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("stock service get by item: %w", err)
	}

	return s.repo.GetByItem(ctx, itemID)
}

// LocationService is a thin read service for locations.
type LocationService struct {
	repo *repository.LocationRepository
}

func NewLocationService(repo *repository.LocationRepository) *LocationService {
	return &LocationService{repo: repo}
}

func (s *LocationService) List(ctx context.Context) ([]model.Location, error) {
	return s.repo.List(ctx)
}
