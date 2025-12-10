package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	bundles_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/bundles"
	sprites_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/sprites"
	bundles_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bundles"
	sprites_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/sprites"
	bundles_service "github.com/FeedTheRealm-org/core-service/internal/assets-service/services/bundles"
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

	/* Model Endpoints */

	bundlesRepo := bundles_repo.NewBundlesRepository(conf, db)
	bundlesService := bundles_service.NewBundlesService(conf, bundlesRepo)
	bundlesController := bundles_controller.NewBundleController(conf, bundlesService)
	bundlesGroup := g.Group("bundles")
	bundlesGroup.GET("/:world_id", bundlesController.DownloadWorldBundle)
	bundlesGroup.POST("/:world_id", bundlesController.UploadWorldBundle)

}
