package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/internal/domain/inventory"
)

// InventoryRepository implements inventory.Repository
type InventoryRepository struct {
	db *pgxpool.Pool
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// Create creates a new inventory record
func (r *InventoryRepository) Create(ctx context.Context, inv *inventory.Inventory) error {
	query := `
		INSERT INTO inventory (
			id, product_id, store_id, quantity, reserved_quantity,
			reorder_point, reorder_quantity, cost_price, selling_price,
			version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		inv.ID, inv.ProductID, inv.StoreID, inv.Quantity, inv.ReservedQuantity,
		inv.ReorderPoint, inv.ReorderQuantity, inv.CostPrice, inv.SellingPrice,
		inv.Version, now, now,
	)

	return err
}

// GetByID retrieves inventory by ID
func (r *InventoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*inventory.Inventory, error) {
	query := `
		SELECT id, product_id, store_id, quantity, reserved_quantity,
			available_quantity, reorder_point, reorder_quantity,
			cost_price, selling_price, version, created_at, updated_at
		FROM inventory
		WHERE id = $1
	`

	var inv inventory.Inventory
	var storeID sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.ProductID, &storeID, &inv.Quantity, &inv.ReservedQuantity,
		&inv.AvailableQuantity, &inv.ReorderPoint, &inv.ReorderQuantity,
		&inv.CostPrice, &inv.SellingPrice, &inv.Version, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if storeID.Valid {
		storeUUID, _ := uuid.Parse(storeID.String)
		inv.StoreID = &storeUUID
	}

	return &inv, nil
}

// GetByProductID retrieves inventory by product ID and optional store ID
func (r *InventoryRepository) GetByProductID(ctx context.Context, productID uuid.UUID, storeID *uuid.UUID) (*inventory.Inventory, error) {
	var query string
	var args []interface{}

	if storeID != nil {
		query = `
			SELECT id, product_id, store_id, quantity, reserved_quantity,
				available_quantity, reorder_point, reorder_quantity,
				cost_price, selling_price, version, created_at, updated_at
			FROM inventory
			WHERE product_id = $1 AND store_id = $2
		`
		args = []interface{}{productID, storeID}
	} else {
		query = `
			SELECT id, product_id, store_id, quantity, reserved_quantity,
				available_quantity, reorder_point, reorder_quantity,
				cost_price, selling_price, version, created_at, updated_at
			FROM inventory
			WHERE product_id = $1 AND store_id IS NULL
		`
		args = []interface{}{productID}
	}

	var inv inventory.Inventory
	var storeIDVal sql.NullString

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&inv.ID, &inv.ProductID, &storeIDVal, &inv.Quantity, &inv.ReservedQuantity,
		&inv.AvailableQuantity, &inv.ReorderPoint, &inv.ReorderQuantity,
		&inv.CostPrice, &inv.SellingPrice, &inv.Version, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if storeIDVal.Valid {
		storeUUID, _ := uuid.Parse(storeIDVal.String)
		inv.StoreID = &storeUUID
	}

	return &inv, nil
}

// GetByStoreID retrieves inventory by store ID
func (r *InventoryRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, limit, offset int) ([]*inventory.Inventory, error) {
	query := `
		SELECT id, product_id, store_id, quantity, reserved_quantity,
			available_quantity, reorder_point, reorder_quantity,
			cost_price, selling_price, version, created_at, updated_at
		FROM inventory
		WHERE store_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, storeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*inventory.Inventory
	for rows.Next() {
		inv, err := scanInventory(rows)
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, inv)
	}

	return inventories, rows.Err()
}

// Update updates inventory
func (r *InventoryRepository) Update(ctx context.Context, inv *inventory.Inventory) error {
	query := `
		UPDATE inventory SET
			quantity = $2, reserved_quantity = $3,
			reorder_point = $4, reorder_quantity = $5,
			cost_price = $6, selling_price = $7,
			version = $8, updated_at = $9
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		inv.ID, inv.Quantity, inv.ReservedQuantity,
		inv.ReorderPoint, inv.ReorderQuantity,
		inv.CostPrice, inv.SellingPrice,
		inv.Version, time.Now(),
	)

	return err
}

// UpdateWithVersion updates inventory with optimistic locking
func (r *InventoryRepository) UpdateWithVersion(ctx context.Context, inv *inventory.Inventory) error {
	query := `
		UPDATE inventory SET
			quantity = $2, reserved_quantity = $3,
			reorder_point = $4, reorder_quantity = $5,
			cost_price = $6, selling_price = $7,
			version = version + 1, updated_at = $8
		WHERE id = $1 AND version = $9
	`

	result, err := r.db.Exec(ctx, query,
		inv.ID, inv.Quantity, inv.ReservedQuantity,
		inv.ReorderPoint, inv.ReorderQuantity,
		inv.CostPrice, inv.SellingPrice,
		time.Now(), inv.Version,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return inventory.ErrVersionConflict
	}

	inv.Version++
	return nil
}

// ReserveStock reserves stock with optimistic locking
func (r *InventoryRepository) ReserveStock(ctx context.Context, productID uuid.UUID, storeID *uuid.UUID, quantity int) error {
	inv, err := r.GetByProductID(ctx, productID, storeID)
	if err != nil {
		return err
	}

	if err := inv.Reserve(quantity); err != nil {
		return err
	}

	return r.UpdateWithVersion(ctx, inv)
}

// ReleaseStock releases reserved stock
func (r *InventoryRepository) ReleaseStock(ctx context.Context, productID uuid.UUID, storeID *uuid.UUID, quantity int) error {
	inv, err := r.GetByProductID(ctx, productID, storeID)
	if err != nil {
		return err
	}

	inv.Release(quantity)
	return r.UpdateWithVersion(ctx, inv)
}

// RecordMovement records a stock movement
func (r *InventoryRepository) RecordMovement(ctx context.Context, movement *inventory.StockMovement) error {
	query := `
		INSERT INTO stock_movements (
			id, inventory_id, movement_type, quantity,
			previous_quantity, new_quantity, reason,
			reference_id, reference_type, user_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		movement.ID, movement.InventoryID, string(movement.MovementType), movement.Quantity,
		movement.PreviousQuantity, movement.NewQuantity, movement.Reason,
		movement.ReferenceID, movement.ReferenceType, movement.UserID, time.Now(),
	)

	return err
}

// GetLowStockItems retrieves items below reorder point
func (r *InventoryRepository) GetLowStockItems(ctx context.Context, storeID *uuid.UUID) ([]*inventory.Inventory, error) {
	var query string
	var args []interface{}

	if storeID != nil {
		query = `
			SELECT id, product_id, store_id, quantity, reserved_quantity,
				available_quantity, reorder_point, reorder_quantity,
				cost_price, selling_price, version, created_at, updated_at
			FROM inventory
			WHERE store_id = $1 AND available_quantity <= reorder_point
			ORDER BY available_quantity ASC
		`
		args = []interface{}{storeID}
	} else {
		query = `
			SELECT id, product_id, store_id, quantity, reserved_quantity,
				available_quantity, reorder_point, reorder_quantity,
				cost_price, selling_price, version, created_at, updated_at
			FROM inventory
			WHERE store_id IS NULL AND available_quantity <= reorder_point
			ORDER BY available_quantity ASC
		`
		args = []interface{}{}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventories []*inventory.Inventory
	for rows.Next() {
		inv, err := scanInventory(rows)
		if err != nil {
			return nil, err
		}
		inventories = append(inventories, inv)
	}

	return inventories, rows.Err()
}

// scanInventory scans a row into an Inventory
func scanInventory(rows interface {
	Scan(dest ...interface{}) error
}) (*inventory.Inventory, error) {
	var inv inventory.Inventory
	var storeID sql.NullString

	err := rows.Scan(
		&inv.ID, &inv.ProductID, &storeID, &inv.Quantity, &inv.ReservedQuantity,
		&inv.AvailableQuantity, &inv.ReorderPoint, &inv.ReorderQuantity,
		&inv.CostPrice, &inv.SellingPrice, &inv.Version, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if storeID.Valid {
		storeUUID, _ := uuid.Parse(storeID.String)
		inv.StoreID = &storeUUID
	}

	return &inv, nil
}

