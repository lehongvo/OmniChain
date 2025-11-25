package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/internal/domain/order"
)

// OrderRepository implements order.Repository
type OrderRepository struct {
	db *pgxpool.Pool
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order
func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}

	var shippingAddrJSON, billingAddrJSON []byte
	if o.ShippingAddress != nil {
		shippingAddrJSON, err = json.Marshal(o.ShippingAddress)
		if err != nil {
			return err
		}
	}
	if o.BillingAddress != nil {
		billingAddrJSON, err = json.Marshal(o.BillingAddress)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO orders (
			id, user_id, store_id, status, total_amount, currency,
			items, shipping_address, billing_address, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	_, err = r.db.Exec(ctx, query,
		o.ID, o.UserID, o.StoreID, string(o.Status), o.TotalAmount, o.Currency,
		itemsJSON, shippingAddrJSON, billingAddrJSON, o.Notes,
		now, now,
	)

	return err
}

// GetByID retrieves an order by ID
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	query := `
		SELECT id, user_id, store_id, status, total_amount, currency,
			items, shipping_address, billing_address, notes,
			created_at, updated_at, completed_at, cancelled_at
		FROM orders
		WHERE id = $1 AND cancelled_at IS NULL
	`

	var o order.Order
	var statusStr string
	var itemsJSON, shippingAddrJSON, billingAddrJSON []byte
	var completedAt, cancelledAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.StoreID, &statusStr, &o.TotalAmount, &o.Currency,
		&itemsJSON, &shippingAddrJSON, &billingAddrJSON, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt, &completedAt, &cancelledAt,
	)
	if err != nil {
		return nil, err
	}

	o.Status = order.OrderStatus(statusStr)
	if completedAt.Valid {
		o.CompletedAt = &completedAt.Time
	}
	if cancelledAt.Valid {
		o.CancelledAt = &cancelledAt.Time
	}

	// Parse JSON fields
	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, err
	}
	if len(shippingAddrJSON) > 0 {
		var addr order.Address
		if err := json.Unmarshal(shippingAddrJSON, &addr); err == nil {
			o.ShippingAddress = &addr
		}
	}
	if len(billingAddrJSON) > 0 {
		var addr order.Address
		if err := json.Unmarshal(billingAddrJSON, &addr); err == nil {
			o.BillingAddress = &addr
		}
	}

	return &o, nil
}

// GetByUserID retrieves orders by user ID
func (r *OrderRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*order.Order, error) {
	query := `
		SELECT id, user_id, store_id, status, total_amount, currency,
			items, shipping_address, billing_address, notes,
			created_at, updated_at, completed_at, cancelled_at
		FROM orders
		WHERE user_id = $1 AND cancelled_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*order.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

// GetByStoreID retrieves orders by store ID
func (r *OrderRepository) GetByStoreID(ctx context.Context, storeID uuid.UUID, limit, offset int) ([]*order.Order, error) {
	query := `
		SELECT id, user_id, store_id, status, total_amount, currency,
			items, shipping_address, billing_address, notes,
			created_at, updated_at, completed_at, cancelled_at
		FROM orders
		WHERE store_id = $1 AND cancelled_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, storeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*order.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

// Update updates an order
func (r *OrderRepository) Update(ctx context.Context, o *order.Order) error {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}

	var shippingAddrJSON, billingAddrJSON []byte
	if o.ShippingAddress != nil {
		shippingAddrJSON, err = json.Marshal(o.ShippingAddress)
		if err != nil {
			return err
		}
	}
	if o.BillingAddress != nil {
		billingAddrJSON, err = json.Marshal(o.BillingAddress)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE orders SET
			status = $2, total_amount = $3, currency = $4,
			items = $5, shipping_address = $6, billing_address = $7,
			notes = $8, updated_at = $9,
			completed_at = $10, cancelled_at = $11
		WHERE id = $1 AND cancelled_at IS NULL
	`

	_, err = r.db.Exec(ctx, query,
		o.ID, string(o.Status), o.TotalAmount, o.Currency,
		itemsJSON, shippingAddrJSON, billingAddrJSON, o.Notes,
		time.Now(), o.CompletedAt, o.CancelledAt,
	)

	return err
}

// UpdateStatus updates order status
func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status order.OrderStatus) error {
	query := `
		UPDATE orders SET
			status = $2,
			updated_at = $3,
			completed_at = CASE WHEN $2 = 'delivered' THEN $3 ELSE completed_at END,
			cancelled_at = CASE WHEN $2 = 'cancelled' THEN $3 ELSE cancelled_at END
		WHERE id = $1 AND cancelled_at IS NULL
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, id, string(status), now)
	return err
}

// Delete soft deletes an order
func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE orders SET
			status = 'cancelled',
			cancelled_at = $2,
			updated_at = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

// CountByUserID counts orders by user ID
func (r *OrderRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE user_id = $1 AND cancelled_at IS NULL`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// scanOrder scans a row into an Order
func scanOrder(rows interface {
	Scan(dest ...interface{}) error
}) (*order.Order, error) {
	var o order.Order
	var statusStr string
	var itemsJSON, shippingAddrJSON, billingAddrJSON []byte
	var completedAt, cancelledAt sql.NullTime

	err := rows.Scan(
		&o.ID, &o.UserID, &o.StoreID, &statusStr, &o.TotalAmount, &o.Currency,
		&itemsJSON, &shippingAddrJSON, &billingAddrJSON, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt, &completedAt, &cancelledAt,
	)
	if err != nil {
		return nil, err
	}

	o.Status = order.OrderStatus(statusStr)
	if completedAt.Valid {
		o.CompletedAt = &completedAt.Time
	}
	if cancelledAt.Valid {
		o.CancelledAt = &cancelledAt.Time
	}

	// Parse JSON fields
	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, err
	}
	if len(shippingAddrJSON) > 0 {
		var addr order.Address
		if err := json.Unmarshal(shippingAddrJSON, &addr); err == nil {
			o.ShippingAddress = &addr
		}
	}
	if len(billingAddrJSON) > 0 {
		var addr order.Address
		if err := json.Unmarshal(billingAddrJSON, &addr); err == nil {
			o.BillingAddress = &addr
		}
	}

	return &o, nil
}

