package notification

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/notification"
	"github.com/onichange/pos-system/pkg/validator"
)

// Handler handles notification HTTP requests
type Handler struct {
	notificationRepo notification.Repository
}

// NewHandler creates a new notification handler
func NewHandler(notificationRepo notification.Repository) *Handler {
	return &Handler{
		notificationRepo: notificationRepo,
	}
}

// GetNotifications handles GET /notifications
func (h *Handler) GetNotifications(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	limit := 20
	offset := 0
	unreadOnly := false

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	if unreadStr := c.Query("unread_only"); unreadStr == "true" {
		unreadOnly = true
	}

	notifications, err := h.notificationRepo.GetByUserID(c.Context(), userID, limit, offset, unreadOnly)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch notifications",
		})
	}

	responses := make([]*NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = ToResponse(n)
	}

	return c.JSON(fiber.Map{
		"data":   responses,
		"limit":  limit,
		"offset": offset,
	})
}

// GetNotification handles GET /notifications/:id
func (h *Handler) GetNotification(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid notification ID",
		})
	}

	n, err := h.notificationRepo.GetByID(c.Context(), notificationID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Notification not found",
		})
	}

	// Check ownership
	if n.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(ToResponse(n))
}

// CreateNotification handles POST /notifications
func (h *Handler) CreateNotification(c *fiber.Ctx) error {
	var req CreateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if validationErrors := validator.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": validationErrors,
		})
	}

	// Parse channels
	channels := make([]notification.Channel, 0)
	if len(req.Channels) > 0 {
		for _, ch := range req.Channels {
			channels = append(channels, notification.Channel(ch))
		}
	} else {
		channels = []notification.Channel{notification.ChannelInApp}
	}

	// Parse priority
	priority := notification.PriorityNormal
	if req.Priority != "" {
		priority = notification.Priority(req.Priority)
	}

	// Parse expires_at
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			expiresAt = &t
		}
	}

	n := &notification.Notification{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Type:      notification.NotificationType(req.Type),
		Title:     req.Title,
		Message:   req.Message,
		Data:      req.Data,
		Channels:  channels,
		Priority:  priority,
		ExpiresAt: expiresAt,
		IsRead:    false,
	}

	if err := h.notificationRepo.Create(c.Context(), n); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create notification",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ToResponse(n))
}

// MarkAsRead handles PUT /notifications/:id/read
func (h *Handler) MarkAsRead(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid notification ID",
		})
	}

	if err := h.notificationRepo.MarkAsRead(c.Context(), notificationID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark notification as read",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead handles PUT /notifications/read-all
func (h *Handler) MarkAllAsRead(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.notificationRepo.MarkAllAsRead(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mark all notifications as read",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "All notifications marked as read",
	})
}

// DeleteNotification handles DELETE /notifications/:id
func (h *Handler) DeleteNotification(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid notification ID",
		})
	}

	if err := h.notificationRepo.Delete(c.Context(), notificationID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete notification",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetUnreadCount handles GET /notifications/unread/count
func (h *Handler) GetUnreadCount(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	count, err := h.notificationRepo.CountUnread(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count unread notifications",
		})
	}

	return c.JSON(fiber.Map{
		"unread_count": count,
	})
}
