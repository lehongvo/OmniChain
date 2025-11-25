package order

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents order status
type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
	StatusRefunded   OrderStatus = "refunded"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
	Discount  float64 `json:"discount,omitempty"`
}

// Address represents shipping/billing address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Order represents an order entity
type Order struct {
	ID              uuid.UUID   `json:"id"`
	UserID          uuid.UUID   `json:"user_id"`
	StoreID         uuid.UUID   `json:"store_id"`
	Status          OrderStatus `json:"status"`
	TotalAmount     float64     `json:"total_amount"`
	Currency        string      `json:"currency"`
	Items           []OrderItem `json:"items"`
	ShippingAddress *Address    `json:"shipping_address,omitempty"`
	BillingAddress  *Address    `json:"billing_address,omitempty"`
	Notes           string      `json:"notes,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	CompletedAt     *time.Time  `json:"completed_at,omitempty"`
	CancelledAt     *time.Time  `json:"cancelled_at,omitempty"`
}

// CalculateTotal calculates total amount from items
func (o *Order) CalculateTotal() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.Subtotal
	}
	return total
}

// CanCancel checks if order can be cancelled
func (o *Order) CanCancel() bool {
	return o.Status == StatusPending || o.Status == StatusConfirmed
}

// CanUpdate checks if order can be updated
func (o *Order) CanUpdate() bool {
	return o.Status == StatusPending || o.Status == StatusConfirmed
}
