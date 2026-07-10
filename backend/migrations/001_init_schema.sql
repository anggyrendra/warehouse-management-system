-- Migration: Create items, locations, and stock tables
-- Description: Initial schema for the Warehouse Item Management API.

-- Items table (master data of warehouse goods)
CREATE TABLE IF NOT EXISTS items (
    id          BIGSERIAL PRIMARY KEY,
    sku         VARCHAR(100) NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL,
    category    VARCHAR(100) NOT NULL,
    unit        VARCHAR(50)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ           DEFAULT NULL
);

-- Partial unique index ensures SKU uniqueness only among non-deleted rows,
-- so soft-deleted SKUs can be reused if needed while active ones stay unique.
CREATE UNIQUE INDEX IF NOT EXISTS uq_items_sku_active
    ON items (sku)
    WHERE deleted_at IS NULL;

-- Locations table (rack/zone in the warehouse)
CREATE TABLE IF NOT EXISTS locations (
    id    BIGSERIAL PRIMARY KEY,
    code  VARCHAR(50)  NOT NULL UNIQUE,
    zone  VARCHAR(50)  NOT NULL,
    type  VARCHAR(50)  NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Stock table (quantity of an item at a given location)
CREATE TABLE IF NOT EXISTS stock (
    id          BIGSERIAL PRIMARY KEY,
    item_id     BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    location_id BIGINT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    qty         INTEGER NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- One stock row per item-location pair
    CONSTRAINT uq_stock_item_location UNIQUE (item_id, location_id),
    -- Database-level guard: qty must never be negative.
    -- Note: validation is ALSO enforced at the service layer per business rules.
    CONSTRAINT chk_stock_qty_nonneg CHECK (qty >= 0)
);

-- Helpful indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_items_category ON items (category);
CREATE INDEX IF NOT EXISTS idx_items_deleted_at ON items (deleted_at);
CREATE INDEX IF NOT EXISTS idx_stock_item_id ON stock (item_id);

-- Seed a few sample locations so stock-receive can be tested immediately
INSERT INTO locations (code, zone, type) VALUES
    ('RACK-A01', 'ZONE-A', 'rack'),
    ('RACK-A02', 'ZONE-A', 'rack'),
    ('RACK-B01', 'ZONE-B', 'rack')
ON CONFLICT (code) DO NOTHING;
