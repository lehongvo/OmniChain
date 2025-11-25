package store

import (
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/store"
)

// CreateStoreRequest represents create store request
type CreateStoreRequest struct {
	Name       string  `json:"name" validate:"required"`
	Code       string  `json:"code" validate:"required"`
	Latitude   float64 `json:"latitude" validate:"required"`
	Longitude  float64 `json:"longitude" validate:"required"`
	Address    string  `json:"address" validate:"required"`
	City       string  `json:"city" validate:"required"`
	State      string  `json:"state" validate:"required"`
	PostalCode string  `json:"postal_code" validate:"required"`
	Country    string  `json:"country" validate:"required"`
	Phone      string  `json:"phone,omitempty"`
	Email      string  `json:"email,omitempty"`
}

// UpdateStoreRequest represents update store request
type UpdateStoreRequest struct {
	Name       string  `json:"name,omitempty"`
	Code       string  `json:"code,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Address    string  `json:"address,omitempty"`
	City       string  `json:"city,omitempty"`
	State      string  `json:"state,omitempty"`
	PostalCode string  `json:"postal_code,omitempty"`
	Country    string  `json:"country,omitempty"`
	Phone      string  `json:"phone,omitempty"`
	Email      string  `json:"email,omitempty"`
	Status     string  `json:"status,omitempty"`
}

// StoreResponse represents store response
type StoreResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Code       string    `json:"code"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Address    string    `json:"address"`
	City       string    `json:"city"`
	State      string    `json:"state"`
	PostalCode string    `json:"postal_code"`
	Country    string    `json:"country"`
	Phone      string    `json:"phone,omitempty"`
	Email      string    `json:"email,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

// ToResponse converts domain Store to StoreResponse
func ToResponse(s *store.Store) *StoreResponse {
	return &StoreResponse{
		ID:         s.ID,
		Name:       s.Name,
		Code:       s.Code,
		Latitude:   s.Latitude,
		Longitude:  s.Longitude,
		Address:    s.Address,
		City:       s.City,
		State:      s.State,
		PostalCode: s.PostalCode,
		Country:    s.Country,
		Phone:      s.Phone,
		Email:      s.Email,
		Status:     string(s.Status),
		CreatedAt:  s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
