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

func SetupEndpointsForItemsService(conf *config.Config, db *config.DB, g *gin.RouterGroup, itemsBucketRepo bucket.BucketRepository) {
	itemsRepo := items_repo.NewItemRepository(conf, db)
	itemsService := items_service.NewItemService(conf, itemsRepo, itemsBucketRepo)
	itemsController := items_controller.NewItemController(conf, itemsService)

	/* Items Endpoints */
	itemsGroup := g.Group("/items")
	itemsGroup.GET("/categories/:id", itemsController.GetItemsListByCategory)
	itemsGroup.GET(":id", itemsController.GetItemById)
	itemsGroup.PUT("/categories/:id", itemsController.UploadItems)

	/* ADMIN ONLY */
	itemsGroup.POST("/categories", itemsController.AddCategory)
}

func SetupEndpointsForModelsService(conf *config.Config, db *config.DB, g *gin.RouterGroup, worldBucketRepo bucket.BucketRepository) {
	modelsRepo := models_repo.NewModelsRepository(conf, db)
	modelsService := models_service.NewModelsService(conf, modelsRepo, worldBucketRepo)
	modelsController := models_controller.NewModelsController(conf, modelsService)

	modelsGroup := g.Group("/models")
	modelsGroup.GET("/world/:world_id", modelsController.GetModelsList)
	modelsGroup.PUT("/world/:world_id", modelsController.UploadModels)
}

func SetupAssetsServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	g := r.Group("/assets")

	cosmeticsBucketRepo, err := bucket.NewOnDiskBucketRepository("cosmetics", conf)
	if err != nil {
		return err
	}

	worldBucketRepo, err := bucket.NewOnDiskBucketRepository("world", conf)
	if err != nil {
		return err
	}

	/* Cosmetics endpoints */
	SetupEndpointsForCosmeticsService(conf, db, g, cosmeticsBucketRepo)

	/* Items endpoints */
	SetupEndpointsForItemsService(conf, db, g, worldBucketRepo)

	// /* Models Endpoints */
	SetupEndpointsForModelsService(conf, db, g, worldBucketRepo)

	return nil
}
