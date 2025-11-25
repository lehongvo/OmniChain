package store

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the store repository interface
type Repository interface {
	Create(ctx context.Context, store *Store) error
	GetByID(ctx context.Context, id uuid.UUID) (*Store, error)
	GetByCode(ctx context.Context, code string) (*Store, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Store, error)
	Update(ctx context.Context, store *Store) error
	Delete(ctx context.Context, id uuid.UUID) error
	SearchByLocation(ctx context.Context, lat, lng float64, radiusKm float64) ([]*Store, error)
}
