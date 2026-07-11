package service

import (
	"context"
	"fmt"
	"time"

	"github.com/anterajatech/warehouse-api/internal/cache"
)

// cacheWrapper adapts the optional Redis cache to the service layer.
// All methods are safe to call regardless of whether caching is enabled —
// they become no-ops when Redis is off.
type cacheWrapper struct {
	c       *cache.Cache
	enabled bool
}

// NewCacheWrapper creates a cacheWrapper from a Cache. Exported so the
// composition root (app.Container) can build it.
func NewCacheWrapper(c *cache.Cache) *cacheWrapper {
	return &cacheWrapper{c: c, enabled: c.Enabled()}
}

const (
	itemTTL   = 5 * time.Minute
	listTTL   = 2 * time.Minute
	stockTTL  = 2 * time.Minute
)

func (cw *cacheWrapper) GetItem(ctx context.Context, key string, dest any) error {
	return cw.c.Get(ctx, key, dest)
}

func (cw *cacheWrapper) SetItem(ctx context.Context, key string, val any) {
	_ = cw.c.Set(ctx, key, val, itemTTL)
}

func (cw *cacheWrapper) SetItemList(ctx context.Context, key string, val any) {
	_ = cw.c.Set(ctx, key, val, listTTL)
}

func (cw *cacheWrapper) InvalidateItem(ctx context.Context, id int64) {
	cw.c.InvalidateItemCache(ctx, id)
}

func (cw *cacheWrapper) InvalidateItemList(ctx context.Context) {
	cw.c.Delete(ctx, "items:list")
}

func (cw *cacheWrapper) Delete(ctx context.Context, key string) {
	cw.c.Delete(ctx, key)
}

// buildStockKey is a small helper to keep cache key construction consistent.
func buildStockKey(itemID int64) string {
	return fmt.Sprintf("stock:%d", itemID)
}
