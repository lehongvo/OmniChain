package inventory

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the inventory repository interface
type Repository interface {
	Create(ctx context.Context, inventory *Inventory) error
	GetByID(ctx context.Context, id uuid.UUID) (*Inventory, error)
	GetByProductID(ctx context.Context, productID uuid.UUID, storeID *uuid.UUID) (*Inventory, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID, limit, offset int) ([]*Inventory, error)
	Update(ctx context.Context, inventory *Inventory) error
	UpdateWithVersion(ctx context.Context, inventory *Inventory) error // Optimistic locking
	ReserveStock(ctx context.Context, productID uuid.UUID, storeID *uuid.UUID, quantity int) error
	ReleaseStock(ctx context.Context, productID uuid.UUID, storeID *uuid.UUID, quantity int) error
	RecordMovement(ctx context.Context, movement *StockMovement) error
	GetLowStockItems(ctx context.Context, storeID *uuid.UUID) ([]*Inventory, error)
}

// StockMovement represents a stock movement record
type StockMovement struct {
	ID              uuid.UUID    `json:"id"`
	InventoryID     uuid.UUID    `json:"inventory_id"`
	MovementType    MovementType `json:"movement_type"`
	Quantity        int          `json:"quantity"`
	PreviousQuantity int         `json:"previous_quantity"`
	NewQuantity     int          `json:"new_quantity"`
	Reason          string       `json:"reason,omitempty"`
	ReferenceID     *uuid.UUID   `json:"reference_id,omitempty"`
	ReferenceType   string       `json:"reference_type,omitempty"`
	UserID          *uuid.UUID   `json:"user_id,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
}

