package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/internal/domain/user"
)

// UserRepository implements user.Repository
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, phone,
			mfa_enabled, mfa_secret, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		u.ID, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Phone,
		u.MFAEnabled, u.MFASecret, now, now,
	)

	return err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone,
			mfa_enabled, mfa_secret, failed_login_attempts, account_locked_until,
			last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var u user.User
	var accountLockedUntil, lastLoginAt, deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.MFAEnabled, &u.MFASecret, &u.FailedLoginAttempts, &accountLockedUntil,
		&lastLoginAt, &u.CreatedAt, &u.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	if accountLockedUntil.Valid {
		u.AccountLockedUntil = &accountLockedUntil.Time
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	if deletedAt.Valid {
		u.DeletedAt = &deletedAt.Time
	}

	return &u, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone,
			mfa_enabled, mfa_secret, failed_login_attempts, account_locked_until,
			last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var u user.User
	var accountLockedUntil, lastLoginAt, deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.MFAEnabled, &u.MFASecret, &u.FailedLoginAttempts, &accountLockedUntil,
		&lastLoginAt, &u.CreatedAt, &u.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	if accountLockedUntil.Valid {
		u.AccountLockedUntil = &accountLockedUntil.Time
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	if deletedAt.Valid {
		u.DeletedAt = &deletedAt.Time
	}

	return &u, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users SET
			first_name = $2, last_name = $3, phone = $4,
			mfa_enabled = $5, mfa_secret = $6,
			failed_login_attempts = $7, account_locked_until = $8,
			last_login_at = $9, updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		u.ID, u.FirstName, u.LastName, u.Phone,
		u.MFAEnabled, u.MFASecret,
		u.FailedLoginAttempts, u.AccountLockedUntil,
		u.LastLoginAt, time.Now(),
	)

	return err
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users SET
			password_hash = $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id, passwordHash, time.Now())
	return err
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users SET
			deleted_at = $2,
			updated_at = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

// ExistsByEmail checks if user exists by email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}
