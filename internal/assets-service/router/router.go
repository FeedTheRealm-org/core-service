package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	cosmetics_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/cosmetics"
	items_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/items"
	models_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	cosmetics_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/cosmetics"
	items_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/items"
	models_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/models"
	cosmetics_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/cosmetics"
	items_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/items"
	models_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/models"
	"github.com/gin-gonic/gin"
)

func SetupAssetsServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	g := r.Group("/assets")

	spritesBucketRepo, err := bucket.NewOnDiskBucketRepository("sprites", conf)
	if err != nil {
		return err
	}

	modelsBucketRepo, err := bucket.NewOnDiskBucketRepository("models", conf)
	if err != nil {
		return err
	}

	spritesRepo := cosmetics_repo.NewSpritesRepository(conf, db)
	spritesService := cosmetics_service.NewSpritesService(conf, spritesRepo, spritesBucketRepo)
	spritesController := cosmetics_controller.NewSpritesController(conf, spritesService)

	spritesGroup := g.Group("/sprites/cosmetics")
	spritesGroup.GET("/categories", spritesController.GetCategoriesList)
	spritesGroup.GET("/categories/:id", spritesController.GetSpritesListByCategory)
	spritesGroup.GET("/:id", spritesController.DownloadSpriteData)

	/* TODO: PROTECT THESE ENDPOINTS: */
	spritesGroup.POST("/categories", spritesController.AddCategory)
	spritesGroup.PUT("", spritesController.UploadSpriteData)

	/* Item Sprites Endpoints */
	itemSpritesRepo := items_repo.NewItemSpritesRepository(conf, db)
	itemSpritesService := items_service.NewItemSpritesService(conf, itemSpritesRepo, spritesBucketRepo)
	itemSpritesController := items_controller.NewItemSpritesController(conf, itemSpritesService)

	itemSpritesGroup := g.Group("/sprites/items")
	itemSpritesGroup.POST(":world_id", itemSpritesController.UploadItemSprite)
	itemSpritesGroup.GET("", itemSpritesController.GetAllItemSprites)
	itemSpritesGroup.GET("/:sprite_id", itemSpritesController.DownloadItemSprite)
	itemSpritesGroup.DELETE("/:sprite_id", itemSpritesController.DeleteItemSprite)

	/* Model Endpoints */

	modelsRepo := models_repo.NewModelsRepository(conf, db)
	modelsService := models_service.NewModelsService(conf, modelsRepo, modelsBucketRepo)
	modelsController := models_controller.NewModelsController(conf, modelsService)
	modelsGroup := g.Group("/models")
	modelsGroup.GET("/:world_id", modelsController.ListAssets)
	modelsGroup.GET("/:world_id/:model_id", modelsController.DownloadModel)
	modelsGroup.POST("/:world_id", modelsController.UploadModels)

	return nil
}
