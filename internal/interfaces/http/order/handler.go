package order

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/order"
	"github.com/onichange/pos-system/pkg/validator"
)

// Handler handles order HTTP requests
type Handler struct {
	orderRepo order.Repository
}

// NewHandler creates a new order handler
func NewHandler(orderRepo order.Repository) *Handler {
	return &Handler{
		orderRepo: orderRepo,
	}
}

// GetOrders handles GET /orders
func (h *Handler) GetOrders(c *fiber.Ctx) error {
	// Get user ID from JWT (set by middleware)
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

	// Parse pagination
	limit := 20
	offset := 0
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

	// Get orders
	orders, err := h.orderRepo.GetByUserID(c.Context(), userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch orders",
		})
	}

	// Convert to response
	responses := make([]*OrderResponse, len(orders))
	for i, o := range orders {
		responses[i] = ToResponse(o)
	}

	return c.JSON(fiber.Map{
		"data":   responses,
		"limit":  limit,
		"offset": offset,
	})
}

// GetOrderByID handles GET /orders/:id
func (h *Handler) GetOrderByID(c *fiber.Ctx) error {
	// Get user ID from JWT
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

	// Parse order ID
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	// Get order
	o, err := h.orderRepo.GetByID(c.Context(), orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// Check ownership
	if o.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(ToResponse(o))
}

// CreateOrder handles POST /orders
func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	// Get user ID from JWT
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

	// Parse request
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if validationErrors := validator.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": validationErrors,
		})
	}

	// Create order
	o := &order.Order{
		ID:              uuid.New(),
		UserID:          userID,
		StoreID:         req.StoreID,
		Status:          order.StatusPending,
		Items:           req.Items,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		Notes:           req.Notes,
		Currency:        "USD",
	}

	// Calculate total
	o.TotalAmount = o.CalculateTotal()

	// Save order
	if err := h.orderRepo.Create(c.Context(), o); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ToResponse(o))
}

// UpdateOrder handles PUT /orders/:id
func (h *Handler) UpdateOrder(c *fiber.Ctx) error {
	// Get user ID from JWT
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

	// Parse order ID
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	// Get existing order
	o, err := h.orderRepo.GetByID(c.Context(), orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// Check ownership
	if o.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Check if can update
	if !o.CanUpdate() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order cannot be updated",
		})
	}

	// Parse request
	var req UpdateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if len(req.Items) > 0 {
		o.Items = req.Items
		o.TotalAmount = o.CalculateTotal()
	}
	if req.ShippingAddress != nil {
		o.ShippingAddress = req.ShippingAddress
	}
	if req.BillingAddress != nil {
		o.BillingAddress = req.BillingAddress
	}
	if req.Notes != "" {
		o.Notes = req.Notes
	}

	// Save order
	if err := h.orderRepo.Update(c.Context(), o); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order",
		})
	}

	return c.JSON(ToResponse(o))
}

// DeleteOrder handles DELETE /orders/:id
func (h *Handler) DeleteOrder(c *fiber.Ctx) error {
	// Get user ID from JWT
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

	// Parse order ID
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	// Get existing order
	o, err := h.orderRepo.GetByID(c.Context(), orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// Check ownership
	if o.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Check if can cancel
	if !o.CanCancel() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order cannot be cancelled",
		})
	}

	// Delete (soft delete)
	if err := h.orderRepo.Delete(c.Context(), orderID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete order",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
