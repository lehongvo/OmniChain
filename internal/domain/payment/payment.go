package payment

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	StatusPending    PaymentStatus = "pending"
	StatusProcessing PaymentStatus = "processing"
	StatusCompleted  PaymentStatus = "completed"
	StatusFailed     PaymentStatus = "failed"
	StatusRefunded   PaymentStatus = "refunded"
)

// PaymentMethodType represents payment method type
type PaymentMethodType string

const (
	MethodCard         PaymentMethodType = "card"
	MethodBankTransfer PaymentMethodType = "bank_transfer"
	MethodDigitalWallet PaymentMethodType = "digital_wallet"
)

// Payment represents a payment entity (PCI-DSS compliant)
type Payment struct {
	ID                  uuid.UUID        `json:"id"`
	OrderID             uuid.UUID        `json:"order_id"`
	UserID              uuid.UUID        `json:"user_id"`
	PaymentMethodToken  string           `json:"payment_method_token"` // Tokenized, never raw card data
	PaymentMethodType   PaymentMethodType `json:"payment_method_type"`
	Amount              float64          `json:"amount"`
	Currency            string           `json:"currency"`
	Status              PaymentStatus    `json:"status"`
	Provider            string           `json:"provider,omitempty"`
	ProviderTransactionID string         `json:"provider_transaction_id,omitempty"`
	ThreeDSecureEnabled bool            `json:"three_d_secure_enabled"`
	ThreeDSecureStatus  string          `json:"three_d_secure_status,omitempty"`
	FraudScore          *float64        `json:"fraud_score,omitempty"`
	FraudFlagged        bool            `json:"fraud_flagged"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	ProcessedAt         *time.Time      `json:"processed_at,omitempty"`
	CompletedAt         *time.Time      `json:"completed_at,omitempty"`
}

// CanRefund checks if payment can be refunded
func (p *Payment) CanRefund() bool {
	return p.Status == StatusCompleted
}

// CanCancel checks if payment can be cancelled
func (p *Payment) CanCancel() bool {
	return p.Status == StatusPending || p.Status == StatusProcessing
}

