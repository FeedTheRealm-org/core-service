package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	itemsprites_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/item-sprites"
	models_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/models"
	sprites_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/sprites"
	itemsprites_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/item-sprites"
	models_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/models"
	sprites_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/sprites"
	itemsprites_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/item-sprites"
	models_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/models"
	sprites_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/sprites"
	"github.com/gin-gonic/gin"
)

func SetupAssetsServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) {
	g := r.Group("/assets")

	spritesRepo := sprites_repo.NewSpritesRepository(conf, db)
	spritesService := sprites_service.NewSpritesService(conf, spritesRepo)
	spritesController := sprites_controller.NewSpritesController(conf, spritesService)

	spritesGroup := g.Group("/sprites")
	spritesGroup.GET("/categories", spritesController.GetCategoriesList)
	spritesGroup.GET("/categories/:id", spritesController.GetSpritesListByCategory)
	spritesGroup.GET("/:id", spritesController.DownloadSpriteData)

	/* TODO: PROTECT THESE ENDPOINTS: */
	spritesGroup.POST("/categories", spritesController.AddCategory)
	spritesGroup.PUT("", spritesController.UploadSpriteData)

	/* Item Sprites Endpoints */
	itemSpritesRepo := itemsprites_repo.NewItemSpritesRepository(conf, db)
	itemSpritesService := itemsprites_service.NewItemSpritesService(conf, itemSpritesRepo)
	itemSpritesController := itemsprites_controller.NewItemSpritesController(conf, itemSpritesService)

	itemSpritesGroup := g.Group("/sprites/items")
	itemSpritesGroup.POST("", itemSpritesController.UploadItemSprite)
	itemSpritesGroup.GET("", itemSpritesController.GetAllItemSprites)
	itemSpritesGroup.GET("/:sprite_id", itemSpritesController.DownloadItemSprite)
	itemSpritesGroup.DELETE("/:sprite_id", itemSpritesController.DeleteItemSprite)
	itemSpritesGroup.GET("/categories", itemSpritesController.GetItemCategories)

	/* Model Endpoints */

	modelsRepo := models_repo.NewModelsRepository(conf, db)
	modelsService := models_service.NewModelsService(conf, modelsRepo)
	modelsController := models_controller.NewModelsController(conf, modelsService)
	modelsGroup := g.Group("/models")
	modelsGroup.GET("/:world_id", modelsController.DownloadModelsByWorldId)
	modelsGroup.POST("", modelsController.UploadModelsByWorldId)

}
