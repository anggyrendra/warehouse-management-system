package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/anterajatech/warehouse-api/internal/middleware"
	"github.com/anterajatech/warehouse-api/internal/model"
	"github.com/anterajatech/warehouse-api/internal/service"
	"github.com/gin-gonic/gin"
)

// ItemHandler handles HTTP requests for the Item resource.
type ItemHandler struct {
	svc *service.ItemService
}

func NewItemHandler(svc *service.ItemService) *ItemHandler {
	return &ItemHandler{svc: svc}
}

// createItemRequest is the validated request body for POST /items.
type createItemRequest struct {
	SKU      string `json:"sku"      binding:"required,min=1,max=100"`
	Name     string `json:"name"     binding:"required,min=1,max=255"`
	Category string `json:"category" binding:"required,min=1,max=100"`
	Unit     string `json:"unit"     binding:"required,min=1,max=50"`
}

// updateItemRequest mirrors createItemRequest but every field is editable.
type updateItemRequest struct {
	SKU      string `json:"sku"      binding:"required,min=1,max=100"`
	Name     string `json:"name"     binding:"required,min=1,max=255"`
	Category string `json:"category" binding:"required,min=1,max=100"`
	Unit     string `json:"unit"     binding:"required,min=1,max=50"`
}

// Create handles POST /api/v1/items
func (h *ItemHandler) Create(c *gin.Context) {
	var req createItemRequest
	// Input validation happens at the handler layer before reaching the service.
	if err := c.ShouldBindJSON(&req); err != nil {
		fields := extractValidationErrors(err)
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    "validation failed",
			Fields: fields,
		})
		return
	}

	item := &model.Item{
		SKU:      strings.TrimSpace(req.SKU),
		Name:     strings.TrimSpace(req.Name),
		Category: strings.TrimSpace(req.Category),
		Unit:     strings.TrimSpace(req.Unit),
	}

	created, err := h.svc.Create(c.Request.Context(), item)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	middleware.Created(c, "Item created successfully", created)
}

// List handles GET /api/v1/items with category filter and pagination.
func (h *ItemHandler) List(c *gin.Context) {
	category := strings.TrimSpace(c.Query("category"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	items, total, err := h.svc.List(c.Request.Context(), category, page, limit)
	if err != nil {
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusInternalServerError,
			Msg:    "failed to retrieve items",
		})
		return
	}

	// Recompute page/limit to reflect clamping done in the service layer.
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	meta := &middleware.Meta{
		Page:  page,
		Limit: limit,
		Total: total,
	}
	middleware.SuccessWithMeta(c, "Items retrieved successfully", items, meta)
}

// Get handles GET /api/v1/items/:id
func (h *ItemHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    "invalid item id",
		})
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	middleware.Success(c, "Item retrieved successfully", item)
}

// Update handles PUT /api/v1/items/:id
func (h *ItemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    "invalid item id",
		})
		return
	}

	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fields := extractValidationErrors(err)
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    "validation failed",
			Fields: fields,
		})
		return
	}

	item := &model.Item{
		ID:       id,
		SKU:      strings.TrimSpace(req.SKU),
		Name:     strings.TrimSpace(req.Name),
		Category: strings.TrimSpace(req.Category),
		Unit:     strings.TrimSpace(req.Unit),
	}

	updated, err := h.svc.Update(c.Request.Context(), item)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	middleware.Success(c, "Item updated successfully", updated)
}

// Delete handles DELETE /api/v1/items/:id (soft delete)
func (h *ItemHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    "invalid item id",
		})
		return
	}

	if err := h.svc.SoftDelete(c.Request.Context(), id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	middleware.Success(c, "Item deleted successfully", nil)
}

// handleServiceError maps service-layer sentinel errors to HTTP responses.
func (h *ItemHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrItemNotFound):
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusNotFound,
			Msg:    err.Error(),
		})
	case errors.Is(err, service.ErrSKUDuplicate):
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusConflict,
			Msg:    err.Error(),
			Fields: []middleware.FieldError{{Field: "sku", Reason: "Duplicate entry"}},
		})
	default:
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusInternalServerError,
			Msg:    "internal server error",
		})
	}
}
