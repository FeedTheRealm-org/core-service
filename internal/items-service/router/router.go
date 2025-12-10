package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	item_controller "github.com/FeedTheRealm-org/core-service/internal/items-service/controllers/item"
	item_repo "github.com/FeedTheRealm-org/core-service/internal/items-service/repositories/item"
	item_service "github.com/FeedTheRealm-org/core-service/internal/items-service/services/item"
	"github.com/gin-gonic/gin"
)

func SetupItemsServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) {
	// Initialize repositories
	itemRepository := item_repo.NewItemRepository(conf, db)
	itemCategoryRepository := item_repo.NewItemCategoryRepository(conf, db)

	// Initialize services
	itemService := item_service.NewItemService(conf, itemRepository)
	itemCategoryService := item_service.NewItemCategoryService(conf, itemCategoryRepository)

	// Initialize controller
	itemController := item_controller.NewItemController(conf, itemService, itemCategoryService)

	// API routes for item metadata
	apiGroup := r.Group("/items")
	{
		apiGroup.POST("", itemController.CreateItem)
		apiGroup.POST("/batch", itemController.CreateItemsBatch)
		apiGroup.GET("/metadata", itemController.GetItemsMetadata)
		apiGroup.GET("/:id", itemController.GetItemById)
		apiGroup.DELETE("/:id", itemController.DeleteItem)

		// Category routes
		apiGroup.POST("/categories", itemController.CreateItemCategory)
		apiGroup.GET("/categories", itemController.GetItemCategories)
		apiGroup.DELETE("/categories/:id", itemController.DeleteItemCategory)

	}
}
