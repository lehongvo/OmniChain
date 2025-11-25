package inventory

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/inventory"
	"github.com/onichange/pos-system/pkg/validator"
)

// Handler handles inventory HTTP requests
type Handler struct {
	inventoryRepo inventory.Repository
}

// NewHandler creates a new inventory handler
func NewHandler(inventoryRepo inventory.Repository) *Handler {
	return &Handler{
		inventoryRepo: inventoryRepo,
	}
}

// GetInventory handles GET /inventory/:id
func (h *Handler) GetInventory(c *fiber.Ctx) error {
	inventoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid inventory ID",
		})
	}

	inv, err := h.inventoryRepo.GetByID(c.Context(), inventoryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Inventory not found",
		})
	}

	return c.JSON(ToResponse(inv))
}

// GetInventoryByProduct handles GET /inventory/product/:product_id
func (h *Handler) GetInventoryByProduct(c *fiber.Ctx) error {
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var storeID *uuid.UUID
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := uuid.Parse(storeIDStr); err == nil {
			storeID = &id
		}
	}

	inv, err := h.inventoryRepo.GetByProductID(c.Context(), productID, storeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Inventory not found",
		})
	}

	return c.JSON(ToResponse(inv))
}

// CreateInventory handles POST /inventory
func (h *Handler) CreateInventory(c *fiber.Ctx) error {
	var req CreateInventoryRequest
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

	inv := &inventory.Inventory{
		ID:               uuid.New(),
		ProductID:        req.ProductID,
		StoreID:          req.StoreID,
		Quantity:         req.Quantity,
		ReservedQuantity: 0,
		ReorderPoint:     req.ReorderPoint,
		ReorderQuantity:  req.ReorderQuantity,
		CostPrice:        req.CostPrice,
		SellingPrice:     req.SellingPrice,
		Version:          1,
	}

	if err := h.inventoryRepo.Create(c.Context(), inv); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create inventory",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ToResponse(inv))
}

// UpdateInventory handles PUT /inventory/:id
func (h *Handler) UpdateInventory(c *fiber.Ctx) error {
	inventoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid inventory ID",
		})
	}

	inv, err := h.inventoryRepo.GetByID(c.Context(), inventoryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Inventory not found",
		})
	}

	var req UpdateInventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if req.Quantity != nil {
		inv.Quantity = *req.Quantity
	}
	if req.ReorderPoint != nil {
		inv.ReorderPoint = *req.ReorderPoint
	}
	if req.ReorderQuantity != nil {
		inv.ReorderQuantity = *req.ReorderQuantity
	}
	if req.CostPrice != nil {
		inv.CostPrice = req.CostPrice
	}
	if req.SellingPrice != nil {
		inv.SellingPrice = req.SellingPrice
	}

	if err := h.inventoryRepo.UpdateWithVersion(c.Context(), inv); err != nil {
		if err == inventory.ErrVersionConflict {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Inventory was modified by another request. Please retry.",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update inventory",
		})
	}

	return c.JSON(ToResponse(inv))
}

// ReserveStock handles POST /inventory/reserve
func (h *Handler) ReserveStock(c *fiber.Ctx) error {
	var req ReserveStockRequest
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

	if err := h.inventoryRepo.ReserveStock(c.Context(), req.ProductID, req.StoreID, req.Quantity); err != nil {
		if err == inventory.ErrInsufficientStock {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Insufficient stock",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reserve stock",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Stock reserved successfully",
	})
}

// ReleaseStock handles POST /inventory/release
func (h *Handler) ReleaseStock(c *fiber.Ctx) error {
	var req ReleaseStockRequest
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

	if err := h.inventoryRepo.ReleaseStock(c.Context(), req.ProductID, req.StoreID, req.Quantity); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to release stock",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Stock released successfully",
	})
}

// GetLowStockItems handles GET /inventory/low-stock
func (h *Handler) GetLowStockItems(c *fiber.Ctx) error {
	var storeID *uuid.UUID
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := uuid.Parse(storeIDStr); err == nil {
			storeID = &id
		}
	}

	items, err := h.inventoryRepo.GetLowStockItems(c.Context(), storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch low stock items",
		})
	}

	responses := make([]*InventoryResponse, len(items))
	for i, inv := range items {
		responses[i] = ToResponse(inv)
	}

	return c.JSON(fiber.Map{"data": responses})
}

// GetInventoryByStore handles GET /inventory/store/:store_id
func (h *Handler) GetInventoryByStore(c *fiber.Ctx) error {
	storeID, err := uuid.Parse(c.Params("store_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid store ID",
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

	items, err := h.inventoryRepo.GetByStoreID(c.Context(), storeID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch inventory",
		})
	}

	responses := make([]*InventoryResponse, len(items))
	for i, inv := range items {
		responses[i] = ToResponse(inv)
	}

	return c.JSON(fiber.Map{
		"data":   responses,
		"limit":  limit,
		"offset": offset,
	})
}
