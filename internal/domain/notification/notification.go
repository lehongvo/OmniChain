package notification

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents notification type
type NotificationType string

const (
	TypeOrder     NotificationType = "order"
	TypePayment   NotificationType = "payment"
	TypeInventory NotificationType = "inventory"
	TypeSystem    NotificationType = "system"
	TypePromotion NotificationType = "promotion"
)

// Priority represents notification priority
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Channel represents delivery channel
type Channel string

const (
	ChannelInApp Channel = "in_app"
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
)

// Notification represents a notification entity
type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      NotificationType `json:"type"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IsRead    bool            `json:"is_read"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	Channels  []Channel       `json:"channels"`
	SentAt    *time.Time      `json:"sent_at,omitempty"`
	Priority  Priority        `json:"priority"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// MarkAsRead marks notification as read
func (n *Notification) MarkAsRead() {
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
}

// IsExpired checks if notification is expired
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

