package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/anterajatech/warehouse-api/internal/middleware"
	"github.com/anterajatech/warehouse-api/internal/model"
	"github.com/anterajatech/warehouse-api/internal/service"
	"github.com/gin-gonic/gin"
)

// StockHandler handles stock-related HTTP requests.
type StockHandler struct {
	svc *service.StockService
}

func NewStockHandler(svc *service.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

// stockReceiveRequest is the validated request body for POST /stock/receive.
type stockReceiveRequest struct {
	ItemID     int64 `json:"item_id"     binding:"required"`
	LocationID int64 `json:"location_id" binding:"required"`
	Qty        int   `json:"qty"         binding:"required,min=1"`
}

// Receive handles POST /api/v1/stock/receive
func (h *StockHandler) Receive(c *gin.Context) {
	var req stockReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fields := extractValidationErrors(err)
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    "validation failed",
			Fields: fields,
		})
		return
	}

	stock, err := h.svc.Receive(c.Request.Context(), model.StockReceive{
		ItemID:     req.ItemID,
		LocationID: req.LocationID,
		Qty:        req.Qty,
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	middleware.Created(c, "Stock received successfully", stock)
}

// GetByItem handles GET /api/v1/stock/:item_id
func (h *StockHandler) GetByItem(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    "invalid item id",
		})
		return
	}

	stocks, err := h.svc.GetByItem(c.Request.Context(), itemID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	middleware.Success(c, "Stock retrieved successfully", stocks)
}

// handleServiceError maps stock service-layer errors to HTTP responses.
func (h *StockHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidStockQty):
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusBadRequest,
			Msg:    err.Error(),
			Fields: []middleware.FieldError{{Field: "qty", Reason: "must be a positive integer"}},
		})
	case errors.Is(err, service.ErrReferencedItem):
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusNotFound,
			Msg:    err.Error(),
			Fields: []middleware.FieldError{{Field: "item_id", Reason: "item not found"}},
		})
	case errors.Is(err, service.ErrReferencedLocation):
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusNotFound,
			Msg:    err.Error(),
			Fields: []middleware.FieldError{{Field: "location_id", Reason: "location not found"}},
		})
	case errors.Is(err, service.ErrItemNotFound):
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusNotFound,
			Msg:    err.Error(),
		})
	default:
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusInternalServerError,
			Msg:    "internal server error",
		})
	}
}
