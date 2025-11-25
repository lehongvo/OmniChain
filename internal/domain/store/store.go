package store

import (
	"time"

	"github.com/google/uuid"
)

// StoreStatus represents store status
type StoreStatus string

const (
	StatusActive   StoreStatus = "active"
	StatusInactive StoreStatus = "inactive"
	StatusClosed   StoreStatus = "closed"
)

// Store represents a store entity
type Store struct {
	ID         uuid.UUID   `json:"id"`
	Name       string      `json:"name"`
	Code       string      `json:"code"`
	Latitude   float64     `json:"latitude"`
	Longitude  float64     `json:"longitude"`
	Address    string      `json:"address"`
	City       string      `json:"city"`
	State      string      `json:"state"`
	PostalCode string      `json:"postal_code"`
	Country    string      `json:"country"`
	Phone      string      `json:"phone,omitempty"`
	Email      string      `json:"email,omitempty"`
	Status     StoreStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
