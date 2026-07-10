package app

import (
	"github.com/anterajatech/warehouse-api/internal/cache"
	"github.com/anterajatech/warehouse-api/internal/config"
	"github.com/anterajatech/warehouse-api/internal/handler"
	"github.com/anterajatech/warehouse-api/internal/middleware"
	"github.com/anterajatech/warehouse-api/internal/repository"
	"github.com/anterajatech/warehouse-api/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Container holds all wired dependencies. Building it once at startup gives
// the application a clear composition root and makes the dependency graph explicit.
type Container struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Cache  *cache.Cache
	Router *gin.Engine
}

// NewContainer bootstraps the entire application: config, database, cache,
// repositories, services, handlers, and the HTTP router with all middleware.
func NewContainer(cfg *config.Config, db *pgxpool.Pool, cch *cache.Cache) *Container {
	// --- Cache wrapper ---
	cw := service.NewCacheWrapper(cch)

	// --- Repositories ---
	itemRepo := repository.NewItemRepository(db)
	locRepo := repository.NewLocationRepository(db)
	stockRepo := repository.NewStockRepository(db)

	// --- Services ---
	itemSvc := service.NewItemService(itemRepo, cw)
	locSvc := service.NewLocationService(locRepo)
	stockSvc := service.NewStockService(stockRepo, itemRepo, locRepo, cw)

	// --- Handlers ---
	itemHandler := handler.NewItemHandler(itemSvc)
	stockHandler := handler.NewStockHandler(stockSvc)
	locHandler := handler.NewLocationHandler(locSvc)
	catHandler := handler.NewCategoriesHandler(itemSvc)
	healthHandler := handler.NewHealthHandler()

	// --- Router setup ---
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Central error handler registered after routes so it can inspect c.Errors.
	r.Use(middleware.ErrorHandler())

	v1 := r.Group("/api/v1")
	{
		// Item routes
		items := v1.Group("/items")
		{
			items.POST("", itemHandler.Create)
			items.GET("", itemHandler.List)
			items.GET("/:id", itemHandler.Get)
			items.PUT("/:id", itemHandler.Update)
			items.DELETE("/:id", itemHandler.Delete)
		}

		// Stock routes
		stock := v1.Group("/stock")
		{
			stock.POST("/receive", stockHandler.Receive)
			stock.GET("/:item_id", stockHandler.GetByItem)
		}

		// Location routes (nilai tambah)
		v1.GET("/locations", locHandler.List)

		// Categories (helper for frontend filter)
		v1.GET("/items-categories", catHandler.List)
	}

	// Health check
	r.GET("/health", healthHandler.Health)

	// Serve OpenAPI/Swagger docs statically from the docs folder.
	r.Static("/docs", "./docs")

	return &Container{Config: cfg, DB: db, Cache: cch, Router: r}
}

// Reference imports to keep the dependency list explicit even if unused directly.
var (
	_ = redis.NewClient
)
