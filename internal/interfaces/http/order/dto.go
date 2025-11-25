package order

import (
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/order"
)

// CreateOrderRequest represents create order request
type CreateOrderRequest struct {
	StoreID         uuid.UUID         `json:"store_id" validate:"required"`
	Items           []order.OrderItem `json:"items" validate:"required,min=1"`
	ShippingAddress *order.Address    `json:"shipping_address,omitempty"`
	BillingAddress  *order.Address    `json:"billing_address,omitempty"`
	Notes           string            `json:"notes,omitempty"`
}

// UpdateOrderRequest represents update order request
type UpdateOrderRequest struct {
	Items           []order.OrderItem `json:"items,omitempty"`
	ShippingAddress *order.Address    `json:"shipping_address,omitempty"`
	BillingAddress  *order.Address    `json:"billing_address,omitempty"`
	Notes           string            `json:"notes,omitempty"`
}

// UpdateOrderStatusRequest represents update order status request
type UpdateOrderStatusRequest struct {
	Status order.OrderStatus `json:"status" validate:"required"`
}

// OrderResponse represents order response
type OrderResponse struct {
	ID              uuid.UUID         `json:"id"`
	UserID          uuid.UUID         `json:"user_id"`
	StoreID         uuid.UUID         `json:"store_id"`
	Status          string            `json:"status"`
	TotalAmount     float64           `json:"total_amount"`
	Currency        string            `json:"currency"`
	Items           []order.OrderItem `json:"items"`
	ShippingAddress *order.Address    `json:"shipping_address,omitempty"`
	BillingAddress  *order.Address    `json:"billing_address,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
	CompletedAt     *string           `json:"completed_at,omitempty"`
	CancelledAt     *string           `json:"cancelled_at,omitempty"`
}

// ToResponse converts domain Order to OrderResponse
func ToResponse(o *order.Order) *OrderResponse {
	resp := &OrderResponse{
		ID:          o.ID,
		UserID:      o.UserID,
		StoreID:     o.StoreID,
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		Currency:    o.Currency,
		Items:       o.Items,
		Notes:       o.Notes,
		CreatedAt:   o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if o.ShippingAddress != nil {
		resp.ShippingAddress = o.ShippingAddress
	}
	if o.BillingAddress != nil {
		resp.BillingAddress = o.BillingAddress
	}
	if o.CompletedAt != nil {
		completedAt := o.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &completedAt
	}
	if o.CancelledAt != nil {
		cancelledAt := o.CancelledAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CancelledAt = &cancelledAt
	}

	return resp
}
