package payment

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/payment"
	"github.com/onichange/pos-system/pkg/validator"
)

// Handler handles payment HTTP requests
type Handler struct {
	paymentRepo payment.Repository
}

// NewHandler creates a new payment handler
func NewHandler(paymentRepo payment.Repository) *Handler {
	return &Handler{
		paymentRepo: paymentRepo,
	}
}

// ProcessPayment handles POST /payments
func (h *Handler) ProcessPayment(c *fiber.Ctx) error {
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

	var req ProcessPaymentRequest
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

	// Create payment (PCI-DSS: only tokenized data, never raw card data)
	p := &payment.Payment{
		ID:                  uuid.New(),
		OrderID:             req.OrderID,
		UserID:              userID,
		PaymentMethodToken:  req.PaymentMethodToken, // Tokenized
		PaymentMethodType:   payment.PaymentMethodType(req.PaymentMethodType),
		Status:              payment.StatusPending,
		Currency:            "USD",
		ThreeDSecureEnabled: req.ThreeDSecure,
		Provider:            "stripe", // Default provider
	}

	// TODO: Integrate with payment provider (Stripe, PayPal, etc.)
	// For now, simulate processing
	p.Status = payment.StatusProcessing
	now := time.Now()
	p.ProcessedAt = &now

	// Save payment
	if err := h.paymentRepo.Create(c.Context(), p); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process payment",
		})
	}

	// Simulate completion (in production, this would be async via webhook)
	go func() {
		time.Sleep(2 * time.Second)
		p.Status = payment.StatusCompleted
		completedAt := time.Now()
		p.CompletedAt = &completedAt
		h.paymentRepo.Update(c.Context(), p)
	}()

	return c.Status(fiber.StatusCreated).JSON(ToResponse(p))
}

// GetPayment handles GET /payments/:id
func (h *Handler) GetPayment(c *fiber.Ctx) error {
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

	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payment ID",
		})
	}

	p, err := h.paymentRepo.GetByID(c.Context(), paymentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Payment not found",
		})
	}

	// Check ownership
	if p.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(ToResponse(p))
}

// GetPaymentsByOrder handles GET /payments/order/:order_id
func (h *Handler) GetPaymentsByOrder(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("order_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	payments, err := h.paymentRepo.GetByOrderID(c.Context(), orderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch payments",
		})
	}

	responses := make([]*PaymentResponse, len(payments))
	for i, p := range payments {
		responses[i] = ToResponse(p)
	}

	return c.JSON(fiber.Map{"data": responses})
}

// GetUserPayments handles GET /payments
func (h *Handler) GetUserPayments(c *fiber.Ctx) error {
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

	payments, err := h.paymentRepo.GetByUserID(c.Context(), userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch payments",
		})
	}

	responses := make([]*PaymentResponse, len(payments))
	for i, p := range payments {
		responses[i] = ToResponse(p)
	}

	return c.JSON(fiber.Map{
		"data":   responses,
		"limit":  limit,
		"offset": offset,
	})
}
