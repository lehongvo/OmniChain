package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/internal/domain/store"
)

// StoreRepository implements store.Repository
type StoreRepository struct {
	db *pgxpool.Pool
}

// NewStoreRepository creates a new store repository
func NewStoreRepository(db *pgxpool.Pool) *StoreRepository {
	return &StoreRepository{db: db}
}

// Create creates a new store
func (r *StoreRepository) Create(ctx context.Context, s *store.Store) error {
	query := `
		INSERT INTO stores (
			id, name, code, latitude, longitude, address, city, state,
			postal_code, country, phone, email, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	now := time.Now()
	isActive := s.Status == store.StatusActive
	_, err := r.db.Exec(ctx, query,
		s.ID, s.Name, s.Code, s.Latitude, s.Longitude, s.Address, s.City, s.State,
		s.PostalCode, s.Country, s.Phone, s.Email, isActive, now, now,
	)

	return err
}

// GetByID retrieves a store by ID
func (r *StoreRepository) GetByID(ctx context.Context, id uuid.UUID) (*store.Store, error) {
	query := `
		SELECT id, name, code, latitude, longitude, address, city, state,
			postal_code, country, phone, email, is_active, created_at, updated_at
		FROM stores
		WHERE id = $1 AND deleted_at IS NULL
	`

	var s store.Store
	var isActive bool

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Code, &s.Latitude, &s.Longitude, &s.Address, &s.City, &s.State,
		&s.PostalCode, &s.Country, &s.Phone, &s.Email, &isActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if isActive {
		s.Status = store.StatusActive
	} else {
		s.Status = store.StatusInactive
	}
	return &s, nil
}

// GetByCode retrieves a store by code
func (r *StoreRepository) GetByCode(ctx context.Context, code string) (*store.Store, error) {
	query := `
		SELECT id, name, code, latitude, longitude, address, city, state,
			postal_code, country, phone, email, is_active, created_at, updated_at
		FROM stores
		WHERE code = $1 AND deleted_at IS NULL
	`

	var s store.Store
	var isActive bool

	err := r.db.QueryRow(ctx, query, code).Scan(
		&s.ID, &s.Name, &s.Code, &s.Latitude, &s.Longitude, &s.Address, &s.City, &s.State,
		&s.PostalCode, &s.Country, &s.Phone, &s.Email, &isActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if isActive {
		s.Status = store.StatusActive
	} else {
		s.Status = store.StatusInactive
	}
	return &s, nil
}

// GetAll retrieves all stores with pagination
func (r *StoreRepository) GetAll(ctx context.Context, limit, offset int) ([]*store.Store, error) {
	query := `
		SELECT id, name, code, latitude, longitude, address, city, state,
			postal_code, country, phone, email, is_active, created_at, updated_at
		FROM stores
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []*store.Store
	for rows.Next() {
		var s store.Store
		var isActive bool

		err := rows.Scan(
			&s.ID, &s.Name, &s.Code, &s.Latitude, &s.Longitude, &s.Address, &s.City, &s.State,
			&s.PostalCode, &s.Country, &s.Phone, &s.Email, &isActive, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if isActive {
			s.Status = store.StatusActive
		} else {
			s.Status = store.StatusInactive
		}
		stores = append(stores, &s)
	}

	return stores, rows.Err()
}

// Update updates a store
func (r *StoreRepository) Update(ctx context.Context, s *store.Store) error {
	query := `
		UPDATE stores SET
			name = $2, code = $3, latitude = $4, longitude = $5,
			address = $6, city = $7, state = $8, postal_code = $9,
			country = $10, phone = $11, email = $12, is_active = $13,
			updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL
	`

	isActive := s.Status == store.StatusActive
	_, err := r.db.Exec(ctx, query,
		s.ID, s.Name, s.Code, s.Latitude, s.Longitude,
		s.Address, s.City, s.State, s.PostalCode,
		s.Country, s.Phone, s.Email, isActive,
		time.Now(),
	)

	return err
}

// Delete soft deletes a store
func (r *StoreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE stores SET
			deleted_at = $2,
			updated_at = $2
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

// SearchByLocation searches stores by location (simplified - using bounding box)
func (r *StoreRepository) SearchByLocation(ctx context.Context, lat, lng float64, radiusKm float64) ([]*store.Store, error) {
	// Simple bounding box search (for production, use PostGIS for accurate distance)
	// Approximate: 1 degree latitude ≈ 111 km
	latDelta := radiusKm / 111.0
	lngDelta := radiusKm / (111.0 * 0.6) // Adjust for longitude

	query := `
		SELECT id, name, code, latitude, longitude, address, city, state,
			postal_code, country, phone, email, is_active, created_at, updated_at
		FROM stores
		WHERE latitude BETWEEN $1 AND $2
			AND longitude BETWEEN $3 AND $4
			AND is_active = TRUE
			AND deleted_at IS NULL
		ORDER BY 
			SQRT(POWER(latitude - $5, 2) + POWER(longitude - $6, 2))
		LIMIT 50
	`

	rows, err := r.db.Query(ctx, query,
		lat-latDelta, lat+latDelta,
		lng-lngDelta, lng+lngDelta,
		lat, lng,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []*store.Store
	for rows.Next() {
		var s store.Store
		var isActive bool

		err := rows.Scan(
			&s.ID, &s.Name, &s.Code, &s.Latitude, &s.Longitude, &s.Address, &s.City, &s.State,
			&s.PostalCode, &s.Country, &s.Phone, &s.Email, &isActive, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if isActive {
			s.Status = store.StatusActive
		} else {
			s.Status = store.StatusInactive
		}
		stores = append(stores, &s)
	}

	return stores, rows.Err()
}

