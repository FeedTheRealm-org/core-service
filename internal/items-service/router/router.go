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
	itemSpriteRepository := item_repo.NewItemSpriteRepository(conf, db)

	// Initialize services
	itemService := item_service.NewItemService(conf, itemRepository)
	itemSpriteService := item_service.NewItemSpriteService(conf, itemSpriteRepository)

	// Initialize controller
	itemController := item_controller.NewItemController(conf, itemService, itemSpriteService)

	// API routes for item metadata
	apiGroup := r.Group("/api/items")
	{
		apiGroup.POST("", itemController.CreateItem)
		apiGroup.POST("/batch", itemController.CreateItemsBatch)
		apiGroup.GET("/metadata", itemController.GetItemsMetadata)
		apiGroup.GET("/:id", itemController.GetItemById)
		apiGroup.DELETE("/:id", itemController.DeleteItem)
	}

	// Assets routes for item sprites
	assetsGroup := r.Group("/assets/sprites/items")
	{
		assetsGroup.POST("", itemController.UploadItemSprite)
		// Download by ID with explicit prefix to avoid route conflict
		assetsGroup.GET("/by-id/:sprite_id", itemController.DownloadItemSprite)
		// Download by category and ID (matches Unity's ItemAssetsService)
		assetsGroup.GET("/:category/:sprite_id", itemController.DownloadItemSpriteByCategory)
		assetsGroup.DELETE("/:sprite_id", itemController.DeleteItemSprite)
	}
}
