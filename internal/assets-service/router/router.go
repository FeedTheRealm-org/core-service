package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	cosmetics_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/cosmetics"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	cosmetics_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/cosmetics"
	cosmetics_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/cosmetics"
	"github.com/gin-gonic/gin"
)

func SetupEndpointsForCosmeticsService(conf *config.Config, db *config.DB, g *gin.RouterGroup, cosmeticsBucketRepo bucket.BucketRepository) {
	cosmeticsRepo := cosmetics_repo.NewCosmeticsRepository(conf, db)
	cosmeticsService := cosmetics_service.NewCosmeticsService(conf, cosmeticsRepo, cosmeticsBucketRepo)
	cosmeticsController := cosmetics_controller.NewCosmeticsController(conf, cosmeticsService)

	/* Cosmetics Endpoints */
	cosmeticsGroup := g.Group("/cosmetics")
	cosmeticsGroup.GET("/categories", cosmeticsController.GetCategoriesList)
	cosmeticsGroup.GET("/categories/:id", cosmeticsController.GetCosmeticsListByCategory)
	cosmeticsGroup.GET(":id", cosmeticsController.GetCosmeticById)
	cosmeticsGroup.PUT("/categories/:id", cosmeticsController.UploadCosmeticData)

	/* ADMIN ONLY */
	cosmeticsGroup.POST("/categories", cosmeticsController.AddCategory)
}

func SetupAssetsServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	g := r.Group("/assets")

	cosmeticsBucketRepo, err := bucket.NewOnDiskBucketRepository("cosmetics", conf)
	if err != nil {
		return err
	}

	// worldBucketRepo, err := bucket.NewOnDiskBucketRepository("world", conf)
	// if err != nil {
	// 	return err
	// }

	/* Cosmetics endpoints */
	SetupEndpointsForCosmeticsService(conf, db, g, cosmeticsBucketRepo)

	/* Items Endpoints */
	// itemsRepo := items_repo.NewItemSpritesRepository(conf, db)
	// itemsService := items_service.NewItemSpritesService(conf, itemsRepo, worldBucketRepo)
	// itemsController := items_controller.NewItemSpritesController(conf, itemsService)

	// itemsGroup := g.Group("/items")
	// itemsGroup.GET("/categories/:id", itemsController.GetItemsListByCategory)
	// itemsGroup.GET(":id", itemsController.GetItemById)
	// itemsGroup.PUT("/categories/:id", itemsController.UploadItems)

	// /* Models Endpoints */
	// modelsRepo := models_repo.NewModelsRepository(conf, db)
	// modelsService := models_service.NewModelsService(conf, modelsRepo, worldBucketRepo)
	// modelsController := models_controller.NewModelsController(conf, modelsService)

	// modelsGroup := g.Group("/models")
	// modelsGroup.GET("/world/:world_id", modelsController.GetModelsList)
	// modelsGroup.PUT("/world/:world_id", modelsController.UploadModels)

	// /* ADMIN ONLY */
	// spritesGroup.POST("/categories", spritesController.AddCategory)
	// itemsGroup.POST("/categories", itemsController.AddCategory)
	// modelsGroup.POST("/world/:world_id", modelsController.AddModel)
	// TODO: DELETE CATEGORIES/SPRITES/MODELS

	return nil
}
