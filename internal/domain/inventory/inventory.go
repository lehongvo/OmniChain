package inventory

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrVersionConflict   = errors.New("version conflict - optimistic locking failed")
)

// MovementType represents stock movement type
type MovementType string

const (
	MovementIn        MovementType = "in"
	MovementOut       MovementType = "out"
	MovementAdjustment MovementType = "adjustment"
	MovementReserved   MovementType = "reserved"
	MovementReleased   MovementType = "released"
)

// Inventory represents inventory entity with optimistic locking
type Inventory struct {
	ID               uuid.UUID `json:"id"`
	ProductID        uuid.UUID `json:"product_id"`
	StoreID          *uuid.UUID `json:"store_id,omitempty"`
	Quantity         int       `json:"quantity"`
	ReservedQuantity int       `json:"reserved_quantity"`
	AvailableQuantity int      `json:"available_quantity"` // Generated column
	ReorderPoint     int       `json:"reorder_point"`
	ReorderQuantity  int       `json:"reorder_quantity"`
	CostPrice        *float64  `json:"cost_price,omitempty"`
	SellingPrice     *float64  `json:"selling_price,omitempty"`
	Version          int       `json:"version"` // For optimistic locking
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Reserve reserves quantity from available stock
func (i *Inventory) Reserve(quantity int) error {
	if i.AvailableQuantity < quantity {
		return ErrInsufficientStock
	}
	i.ReservedQuantity += quantity
	i.Version++
	return nil
}

// Release releases reserved quantity
func (i *Inventory) Release(quantity int) {
	if i.ReservedQuantity >= quantity {
		i.ReservedQuantity -= quantity
		i.Version++
	}
}

// Add adds quantity to inventory
func (i *Inventory) Add(quantity int) {
	i.Quantity += quantity
	i.Version++
}

// Subtract subtracts quantity from inventory
func (i *Inventory) Subtract(quantity int) error {
	if i.AvailableQuantity < quantity {
		return ErrInsufficientStock
	}
	i.Quantity -= quantity
	i.Version++
	return nil
}

// NeedsReorder checks if inventory needs reorder
func (i *Inventory) NeedsReorder() bool {
	return i.AvailableQuantity <= i.ReorderPoint
}

