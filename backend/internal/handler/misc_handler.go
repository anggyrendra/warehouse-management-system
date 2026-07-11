package handler

import (
	"errors"
	"net/http"

	"github.com/anterajatech/warehouse-api/internal/middleware"
	"github.com/anterajatech/warehouse-api/internal/service"
	"github.com/gin-gonic/gin"
)

// LocationHandler handles location-related HTTP requests (nilai tambah).
type LocationHandler struct {
	svc *service.LocationService
}

func NewLocationHandler(svc *service.LocationService) *LocationHandler {
	return &LocationHandler{svc: svc}
}

// List handles GET /api/v1/locations
func (h *LocationHandler) List(c *gin.Context) {
	locs, err := h.svc.List(c.Request.Context())
	if err != nil {
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusInternalServerError,
			Msg:    "failed to retrieve locations",
		})
		return
	}
	middleware.Success(c, "Locations retrieved successfully", locs)
}

// CategoriesHandler exposes distinct item categories for the frontend filter.
type CategoriesHandler struct {
	svc *service.ItemService
}

func NewCategoriesHandler(svc *service.ItemService) *CategoriesHandler {
	return &CategoriesHandler{svc: svc}
}

func (h *CategoriesHandler) List(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		middleware.AbortWithError(c, middleware.ErrorResult{
			Status: http.StatusInternalServerError,
			Msg:    "failed to retrieve categories",
		})
		return
	}
	middleware.Success(c, "Categories retrieved successfully", cats)
}

// HealthHandler returns a simple health-check response.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

func (h *HealthHandler) Health(c *gin.Context) {
	middleware.Success(c, "service is healthy", gin.H{"status": "ok"})
}

// ensure errors import is used (for potential future extension)
var _ = errors.New
