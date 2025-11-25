package notification

import (
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/notification"
)

// CreateNotificationRequest represents create notification request
type CreateNotificationRequest struct {
	UserID    uuid.UUID              `json:"user_id" validate:"required"`
	Type      string                 `json:"type" validate:"required"`
	Title     string                 `json:"title" validate:"required"`
	Message   string                 `json:"message" validate:"required"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Channels  []string               `json:"channels,omitempty"`
	Priority  string                 `json:"priority,omitempty"`
	ExpiresAt *string                `json:"expires_at,omitempty"`
}

// NotificationResponse represents notification response
type NotificationResponse struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IsRead    bool                   `json:"is_read"`
	ReadAt    *string                `json:"read_at,omitempty"`
	Channels  []string               `json:"channels"`
	SentAt    *string                `json:"sent_at,omitempty"`
	Priority  string                 `json:"priority"`
	ExpiresAt *string                `json:"expires_at,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

// ToResponse converts domain Notification to NotificationResponse
func ToResponse(n *notification.Notification) *NotificationResponse {
	resp := &NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      string(n.Type),
		Title:     n.Title,
		Message:   n.Message,
		Data:      n.Data,
		IsRead:    n.IsRead,
		Priority:  string(n.Priority),
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	channels := make([]string, len(n.Channels))
	for i, ch := range n.Channels {
		channels[i] = string(ch)
	}
	resp.Channels = channels

	if n.ReadAt != nil {
		readAt := n.ReadAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ReadAt = &readAt
	}
	if n.SentAt != nil {
		sentAt := n.SentAt.Format("2006-01-02T15:04:05Z07:00")
		resp.SentAt = &sentAt
	}
	if n.ExpiresAt != nil {
		expiresAt := n.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ExpiresAt = &expiresAt
	}

	return resp
}
