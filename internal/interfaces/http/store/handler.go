package store

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/onichange/pos-system/internal/domain/store"
	"github.com/onichange/pos-system/pkg/validator"
)

// Handler handles store HTTP requests
type Handler struct {
	storeRepo store.Repository
}

// NewHandler creates a new store handler
func NewHandler(storeRepo store.Repository) *Handler {
	return &Handler{
		storeRepo: storeRepo,
	}
}

// GetStores handles GET /stores
func (h *Handler) GetStores(c *fiber.Ctx) error {
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

	stores, err := h.storeRepo.GetAll(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stores",
		})
	}

	responses := make([]*StoreResponse, len(stores))
	for i, s := range stores {
		responses[i] = ToResponse(s)
	}

	return c.JSON(fiber.Map{
		"data":   responses,
		"limit":  limit,
		"offset": offset,
	})
}

// GetStoreByID handles GET /stores/:id
func (h *Handler) GetStoreByID(c *fiber.Ctx) error {
	storeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid store ID",
		})
	}

	s, err := h.storeRepo.GetByID(c.Context(), storeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Store not found",
		})
	}

	return c.JSON(ToResponse(s))
}

// CreateStore handles POST /stores
func (h *Handler) CreateStore(c *fiber.Ctx) error {
	var req CreateStoreRequest
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

	s := &store.Store{
		ID:         uuid.New(),
		Name:       req.Name,
		Code:       req.Code,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Address:    req.Address,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		Phone:      req.Phone,
		Email:      req.Email,
		Status:     store.StatusActive,
	}

	if err := h.storeRepo.Create(c.Context(), s); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create store",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(ToResponse(s))
}

// UpdateStore handles PUT /stores/:id
func (h *Handler) UpdateStore(c *fiber.Ctx) error {
	storeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid store ID",
		})
	}

	s, err := h.storeRepo.GetByID(c.Context(), storeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Store not found",
		})
	}

	var req UpdateStoreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update fields
	if req.Name != "" {
		s.Name = req.Name
	}
	if req.Code != "" {
		s.Code = req.Code
	}
	if req.Latitude != 0 {
		s.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		s.Longitude = req.Longitude
	}
	if req.Address != "" {
		s.Address = req.Address
	}
	if req.City != "" {
		s.City = req.City
	}
	if req.State != "" {
		s.State = req.State
	}
	if req.PostalCode != "" {
		s.PostalCode = req.PostalCode
	}
	if req.Country != "" {
		s.Country = req.Country
	}
	if req.Phone != "" {
		s.Phone = req.Phone
	}
	if req.Email != "" {
		s.Email = req.Email
	}
	if req.Status != "" {
		s.Status = store.StoreStatus(req.Status)
	}

	if err := h.storeRepo.Update(c.Context(), s); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update store",
		})
	}

	return c.JSON(ToResponse(s))
}

// DeleteStore handles DELETE /stores/:id
func (h *Handler) DeleteStore(c *fiber.Ctx) error {
	storeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid store ID",
		})
	}

	if err := h.storeRepo.Delete(c.Context(), storeID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete store",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// SearchStores handles GET /stores/search?lat=&lng=&radius=
func (h *Handler) SearchStores(c *fiber.Ctx) error {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.Query("radius", "10") // Default 10km

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid latitude",
		})
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid longitude",
		})
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid radius",
		})
	}

	stores, err := h.storeRepo.SearchByLocation(c.Context(), lat, lng, radius)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search stores",
		})
	}

	responses := make([]*StoreResponse, len(stores))
	for i, s := range stores {
		responses[i] = ToResponse(s)
	}

	return c.JSON(fiber.Map{
		"data": responses,
	})
}
