package model

import (
	"time"
)

// Item represents a master-data row for a warehouse good.
type Item struct {
	ID        int64      `json:"id"`
	SKU       string     `json:"sku"`
	Name      string     `json:"name"`
	Category  string     `json:"category"`
	Unit      string     `json:"unit"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// Location represents a rack/zone inside the warehouse.
type Location struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Zone      string    `json:"zone"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// Stock represents the quantity of a specific item at a specific location.
type Stock struct {
	ID         int64     `json:"id"`
	ItemID     int64     `json:"item_id"`
	LocationID int64     `json:"location_id"`
	Qty        int       `json:"qty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// StockWithDetail is the read model returned by the stock-check endpoint,
// joining stock rows with their parent item and location.
type StockWithDetail struct {
	ID         int64  `json:"id"`
	ItemID     int64  `json:"item_id"`
	ItemSKU    string `json:"item_sku"`
	ItemName   string `json:"item_name"`
	LocationID int64  `json:"location_id"`
	LocationCode string `json:"location_code"`
	Zone       string `json:"zone"`
	Qty        int    `json:"qty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// StockReceive is the request payload for the stock-receive endpoint.
type StockReceive struct {
	ItemID     int64 `json:"item_id"`
	LocationID int64 `json:"location_id"`
	Qty        int   `json:"qty"`
}

// PaginatedItems is the query result returned from the repository.
type PaginatedItems struct {
	Items []Item
	Total int64
}
