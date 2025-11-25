package payment

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the payment repository interface
type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*Payment, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Payment, error)
	Update(ctx context.Context, payment *Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error
}
