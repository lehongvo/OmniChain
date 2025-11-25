package order

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the order repository interface
type Repository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Order, error)
	GetByStoreID(ctx context.Context, storeID uuid.UUID, limit, offset int) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}
