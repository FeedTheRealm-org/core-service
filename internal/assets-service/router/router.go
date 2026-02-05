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

	/* Cosmetics endpoints */

	spritesRepo := cosmetics_repo.NewSpritesRepository(conf, db)
	spritesService := cosmetics_service.NewSpritesService(conf, spritesRepo, spritesBucketRepo)
	spritesController := cosmetics_controller.NewSpritesController(conf, spritesService)

	spritesGroup := g.Group("/cosmetics")
	spritesGroup.GET("/categories", spritesController.GetCategoriesList)
	spritesGroup.GET("/categories/:id", spritesController.GetCosmeticsListByCategory)
	spritesGroup.GET(":id", spritesController.GetCosmeticById)
	spritesGroup.PUT("/categories/:id", spritesController.UploadCosmetics)

	/* Items Endpoints */
	itemsRepo := items_repo.NewItemSpritesRepository(conf, db)
	itemsService := items_service.NewItemSpritesService(conf, itemsRepo, spritesBucketRepo)
	itemsController := items_controller.NewItemSpritesController(conf, itemsService)

	itemsGroup := g.Group("/items")
	spritesGroup.GET("/categories", spritesController.GetCategoriesList)
	itemsGroup.GET("/categories/:id", itemsController.GetItemsListByCategory)
	itemsGroup.GET(":id", itemsController.GetItemById)
	itemsGroup.PUT("/categories/:id", itemsController.UploadItems)

	/* Models Endpoints */

	modelsRepo := models_repo.NewModelsRepository(conf, db)
	modelsService := models_service.NewModelsService(conf, modelsRepo, modelsBucketRepo)
	modelsController := models_controller.NewModelsController(conf, modelsService)

	modelsGroup := g.Group("/models")
	modelsGroup.GET("/world/:world_id", modelsController.GetModelsList)
	modelsGroup.PUT("/world/:world_id", modelsController.UploadModels)

	/* ADMIN ONLY */
	spritesGroup.POST("/categories", spritesController.AddCategory)
	itemsGroup.POST("/categories", itemsGroup.AddCategory)
	// TODO: DELETE CATEGORIES/SPRITES/MODELS

	return nil
}
