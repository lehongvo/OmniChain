package inventory

import (
	"github.com/google/uuid"
	"github.com/onichange/pos-system/internal/domain/inventory"
)

// CreateInventoryRequest represents create inventory request
type CreateInventoryRequest struct {
	ProductID    uuid.UUID  `json:"product_id" validate:"required"`
	StoreID      *uuid.UUID `json:"store_id,omitempty"`
	Quantity     int        `json:"quantity" validate:"min=0"`
	ReorderPoint int        `json:"reorder_point" validate:"min=0"`
	ReorderQuantity int     `json:"reorder_quantity" validate:"min=0"`
	CostPrice    *float64   `json:"cost_price,omitempty"`
	SellingPrice *float64   `json:"selling_price,omitempty"`
}

// UpdateInventoryRequest represents update inventory request
type UpdateInventoryRequest struct {
	Quantity     *int       `json:"quantity,omitempty"`
	ReorderPoint *int       `json:"reorder_point,omitempty"`
	ReorderQuantity *int    `json:"reorder_quantity,omitempty"`
	CostPrice    *float64   `json:"cost_price,omitempty"`
	SellingPrice *float64   `json:"selling_price,omitempty"`
}

// ReserveStockRequest represents reserve stock request
type ReserveStockRequest struct {
	ProductID uuid.UUID  `json:"product_id" validate:"required"`
	StoreID   *uuid.UUID `json:"store_id,omitempty"`
	Quantity  int        `json:"quantity" validate:"required,min=1"`
	Reason    string     `json:"reason,omitempty"`
}

// ReleaseStockRequest represents release stock request
type ReleaseStockRequest struct {
	ProductID uuid.UUID  `json:"product_id" validate:"required"`
	StoreID   *uuid.UUID `json:"store_id,omitempty"`
	Quantity  int        `json:"quantity" validate:"required,min=1"`
}

// InventoryResponse represents inventory response
type InventoryResponse struct {
	ID               uuid.UUID  `json:"id"`
	ProductID        uuid.UUID  `json:"product_id"`
	StoreID          *uuid.UUID `json:"store_id,omitempty"`
	Quantity         int        `json:"quantity"`
	ReservedQuantity int        `json:"reserved_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	ReorderPoint     int        `json:"reorder_point"`
	ReorderQuantity  int        `json:"reorder_quantity"`
	CostPrice        *float64   `json:"cost_price,omitempty"`
	SellingPrice     *float64   `json:"selling_price,omitempty"`
	Version          int        `json:"version"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
}

// ToResponse converts domain Inventory to InventoryResponse
func ToResponse(inv *inventory.Inventory) *InventoryResponse {
	return &InventoryResponse{
		ID:                inv.ID,
		ProductID:         inv.ProductID,
		StoreID:           inv.StoreID,
		Quantity:          inv.Quantity,
		ReservedQuantity:  inv.ReservedQuantity,
		AvailableQuantity: inv.AvailableQuantity,
		ReorderPoint:      inv.ReorderPoint,
		ReorderQuantity:   inv.ReorderQuantity,
		CostPrice:         inv.CostPrice,
		SellingPrice:      inv.SellingPrice,
		Version:           inv.Version,
		CreatedAt:         inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         inv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

