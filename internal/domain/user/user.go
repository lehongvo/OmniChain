package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity
type User struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	PasswordHash      string     `json:"-"` // Never expose in JSON
	FirstName         string     `json:"first_name,omitempty"`
	LastName          string     `json:"last_name,omitempty"`
	Phone             string     `json:"phone,omitempty"`
	MFAEnabled        bool       `json:"mfa_enabled"`
	MFASecret         string     `json:"-"` // Never expose
	FailedLoginAttempts int      `json:"-"`
	AccountLockedUntil *time.Time `json:"-"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"-"`
}

// IsLocked checks if account is locked
func (u *User) IsLocked() bool {
	if u.AccountLockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.AccountLockedUntil)
}

// LockAccount locks the account for specified duration
func (u *User) LockAccount(duration time.Duration) {
	lockedUntil := time.Now().Add(duration)
	u.AccountLockedUntil = &lockedUntil
}

// UnlockAccount unlocks the account
func (u *User) UnlockAccount() {
	u.AccountLockedUntil = nil
	u.FailedLoginAttempts = 0
}

// IncrementFailedLogin increments failed login attempts
func (u *User) IncrementFailedLogin() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= 5 {
		u.LockAccount(30 * time.Minute) // Lock for 30 minutes
	}
}

// ResetFailedLogin resets failed login attempts
func (u *User) ResetFailedLogin() {
	u.FailedLoginAttempts = 0
}

// UpdateLastLogin updates last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

