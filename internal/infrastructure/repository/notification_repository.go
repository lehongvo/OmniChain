package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/onichange/pos-system/internal/domain/notification"
)

// NotificationRepository implements notification.Repository
type NotificationRepository struct {
	db *pgxpool.Pool
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, type, title, message, data,
			channels, priority, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	dataJSON, _ := json.Marshal(n.Data)
	channels := make([]string, len(n.Channels))
	for i, ch := range n.Channels {
		channels[i] = string(ch)
	}

	_, err := r.db.Exec(ctx, query,
		n.ID, n.UserID, string(n.Type), n.Title, n.Message, dataJSON,
		channels, string(n.Priority), n.ExpiresAt, time.Now(),
	)

	return err
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*notification.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, data,
			is_read, read_at, channels, sent_at, priority,
			expires_at, created_at
		FROM notifications
		WHERE id = $1
	`

	var n notification.Notification
	var typeStr, priorityStr string
	var dataJSON []byte
	var channels []string
	var readAt, sentAt, expiresAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.UserID, &typeStr, &n.Title, &n.Message, &dataJSON,
		&n.IsRead, &readAt, &channels, &sentAt, &priorityStr,
		&expiresAt, &n.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	n.Type = notification.NotificationType(typeStr)
	n.Priority = notification.Priority(priorityStr)
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}
	if sentAt.Valid {
		n.SentAt = &sentAt.Time
	}
	if expiresAt.Valid {
		n.ExpiresAt = &expiresAt.Time
	}

	// Parse JSON data
	if len(dataJSON) > 0 {
		json.Unmarshal(dataJSON, &n.Data)
	}

	// Parse channels
	n.Channels = make([]notification.Channel, len(channels))
	for i, ch := range channels {
		n.Channels[i] = notification.Channel(ch)
	}

	return &n, nil
}

// GetByUserID retrieves notifications by user ID
func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, unreadOnly bool) ([]*notification.Notification, error) {
	var query string
	if unreadOnly {
		query = `
			SELECT id, user_id, type, title, message, data,
				is_read, read_at, channels, sent_at, priority,
				expires_at, created_at
			FROM notifications
			WHERE user_id = $1 AND is_read = FALSE
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
	} else {
		query = `
			SELECT id, user_id, type, title, message, data,
				is_read, read_at, channels, sent_at, priority,
				expires_at, created_at
			FROM notifications
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
	}

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE notifications SET
			is_read = TRUE,
			read_at = $3
		WHERE id = $1 AND user_id = $2
	`

	_, err := r.db.Exec(ctx, query, id, userID, time.Now())
	return err
}

// MarkAllAsRead marks all user notifications as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications SET
			is_read = TRUE,
			read_at = $2
		WHERE user_id = $1 AND is_read = FALSE
	`

	_, err := r.db.Exec(ctx, query, userID, time.Now())
	return err
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, id, userID)
	return err
}

// CountUnread counts unread notifications for a user
func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// scanNotification scans a row into a Notification
func scanNotification(rows interface {
	Scan(dest ...interface{}) error
}) (*notification.Notification, error) {
	var n notification.Notification
	var typeStr, priorityStr string
	var dataJSON []byte
	var channels []string
	var readAt, sentAt, expiresAt sql.NullTime

	err := rows.Scan(
		&n.ID, &n.UserID, &typeStr, &n.Title, &n.Message, &dataJSON,
		&n.IsRead, &readAt, &channels, &sentAt, &priorityStr,
		&expiresAt, &n.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	n.Type = notification.NotificationType(typeStr)
	n.Priority = notification.Priority(priorityStr)
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}
	if sentAt.Valid {
		n.SentAt = &sentAt.Time
	}
	if expiresAt.Valid {
		n.ExpiresAt = &expiresAt.Time
	}

	// Parse JSON data
	if len(dataJSON) > 0 {
		json.Unmarshal(dataJSON, &n.Data)
	}

	// Parse channels
	n.Channels = make([]notification.Channel, len(channels))
	for i, ch := range channels {
		n.Channels[i] = notification.Channel(ch)
	}

	return &n, nil
}

