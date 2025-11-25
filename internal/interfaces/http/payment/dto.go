package payment

import (
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/payment"
)

// ProcessPaymentRequest represents process payment request
type ProcessPaymentRequest struct {
	OrderID            uuid.UUID `json:"order_id" validate:"required"`
	PaymentMethodToken string    `json:"payment_method_token" validate:"required"`
	PaymentMethodType  string    `json:"payment_method_type" validate:"required"`
	ThreeDSecure       bool      `json:"three_d_secure,omitempty"`
}

// PaymentResponse represents payment response
type PaymentResponse struct {
	ID                    uuid.UUID `json:"id"`
	OrderID               uuid.UUID `json:"order_id"`
	UserID                uuid.UUID `json:"user_id"`
	PaymentMethodType     string    `json:"payment_method_type"`
	Amount                float64   `json:"amount"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	Provider              string    `json:"provider,omitempty"`
	ProviderTransactionID string    `json:"provider_transaction_id,omitempty"`
	ThreeDSecureEnabled   bool      `json:"three_d_secure_enabled"`
	CreatedAt             string    `json:"created_at"`
	UpdatedAt             string    `json:"updated_at"`
	ProcessedAt           *string   `json:"processed_at,omitempty"`
	CompletedAt           *string   `json:"completed_at,omitempty"`
}

// ToResponse converts domain Payment to PaymentResponse
func ToResponse(p *payment.Payment) *PaymentResponse {
	resp := &PaymentResponse{
		ID:                    p.ID,
		OrderID:               p.OrderID,
		UserID:                p.UserID,
		PaymentMethodType:     string(p.PaymentMethodType),
		Amount:                p.Amount,
		Currency:              p.Currency,
		Status:                string(p.Status),
		Provider:              p.Provider,
		ProviderTransactionID: p.ProviderTransactionID,
		ThreeDSecureEnabled:   p.ThreeDSecureEnabled,
		CreatedAt:             p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:             p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.ProcessedAt != nil {
		processedAt := p.ProcessedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ProcessedAt = &processedAt
	}
	if p.CompletedAt != nil {
		completedAt := p.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &completedAt
	}

	return resp
}
