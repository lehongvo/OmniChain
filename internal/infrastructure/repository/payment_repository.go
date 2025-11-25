package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/internal/domain/payment"
)

// PaymentRepository implements payment.Repository
type PaymentRepository struct {
	db *pgxpool.Pool
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment
func (r *PaymentRepository) Create(ctx context.Context, p *payment.Payment) error {
	query := `
		INSERT INTO payments (
			id, order_id, user_id, payment_method_token, payment_method_type,
			amount, currency, status, provider, provider_transaction_id,
			three_d_secure_enabled, three_d_secure_status, fraud_score, fraud_flagged,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		p.ID, p.OrderID, p.UserID, p.PaymentMethodToken, string(p.PaymentMethodType),
		p.Amount, p.Currency, string(p.Status), p.Provider, p.ProviderTransactionID,
		p.ThreeDSecureEnabled, p.ThreeDSecureStatus, p.FraudScore, p.FraudFlagged,
		now, now,
	)

	return err
}

// GetByID retrieves a payment by ID
func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	query := `
		SELECT id, order_id, user_id, payment_method_token, payment_method_type,
			amount, currency, status, provider, provider_transaction_id,
			three_d_secure_enabled, three_d_secure_status, fraud_score, fraud_flagged,
			created_at, updated_at, processed_at, completed_at
		FROM payments
		WHERE id = $1
	`

	var p payment.Payment
	var methodTypeStr, statusStr string
	var processedAt, completedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.PaymentMethodToken, &methodTypeStr,
		&p.Amount, &p.Currency, &statusStr, &p.Provider, &p.ProviderTransactionID,
		&p.ThreeDSecureEnabled, &p.ThreeDSecureStatus, &p.FraudScore, &p.FraudFlagged,
		&p.CreatedAt, &p.UpdatedAt, &processedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	p.PaymentMethodType = payment.PaymentMethodType(methodTypeStr)
	p.Status = payment.PaymentStatus(statusStr)
	if processedAt.Valid {
		p.ProcessedAt = &processedAt.Time
	}
	if completedAt.Valid {
		p.CompletedAt = &completedAt.Time
	}

	return &p, nil
}

// GetByOrderID retrieves payments by order ID
func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*payment.Payment, error) {
	query := `
		SELECT id, order_id, user_id, payment_method_token, payment_method_type,
			amount, currency, status, provider, provider_transaction_id,
			three_d_secure_enabled, three_d_secure_status, fraud_score, fraud_flagged,
			created_at, updated_at, processed_at, completed_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*payment.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	return payments, rows.Err()
}

// GetByUserID retrieves payments by user ID
func (r *PaymentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*payment.Payment, error) {
	query := `
		SELECT id, order_id, user_id, payment_method_token, payment_method_type,
			amount, currency, status, provider, provider_transaction_id,
			three_d_secure_enabled, three_d_secure_status, fraud_score, fraud_flagged,
			created_at, updated_at, processed_at, completed_at
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*payment.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	return payments, rows.Err()
}

// Update updates a payment
func (r *PaymentRepository) Update(ctx context.Context, p *payment.Payment) error {
	query := `
		UPDATE payments SET
			status = $2, provider = $3, provider_transaction_id = $4,
			three_d_secure_status = $5, fraud_score = $6, fraud_flagged = $7,
			processed_at = $8, completed_at = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		p.ID, string(p.Status), p.Provider, p.ProviderTransactionID,
		p.ThreeDSecureStatus, p.FraudScore, p.FraudFlagged,
		p.ProcessedAt, p.CompletedAt, time.Now(),
	)

	return err
}

// UpdateStatus updates payment status
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status payment.PaymentStatus) error {
	query := `
		UPDATE payments SET
			status = $2,
			updated_at = $3,
			processed_at = CASE WHEN $2 = 'processing' THEN $3 ELSE processed_at END,
			completed_at = CASE WHEN $2 = 'completed' THEN $3 ELSE completed_at END
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, id, string(status), now)
	return err
}

// scanPayment scans a row into a Payment
func scanPayment(rows interface {
	Scan(dest ...interface{}) error
}) (*payment.Payment, error) {
	var p payment.Payment
	var methodTypeStr, statusStr string
	var processedAt, completedAt sql.NullTime

	err := rows.Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.PaymentMethodToken, &methodTypeStr,
		&p.Amount, &p.Currency, &statusStr, &p.Provider, &p.ProviderTransactionID,
		&p.ThreeDSecureEnabled, &p.ThreeDSecureStatus, &p.FraudScore, &p.FraudFlagged,
		&p.CreatedAt, &p.UpdatedAt, &processedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	p.PaymentMethodType = payment.PaymentMethodType(methodTypeStr)
	p.Status = payment.PaymentStatus(statusStr)
	if processedAt.Valid {
		p.ProcessedAt = &processedAt.Time
	}
	if completedAt.Valid {
		p.CompletedAt = &completedAt.Time
	}

	return &p, nil
}
