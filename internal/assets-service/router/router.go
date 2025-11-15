package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	sprites_controller "github.com/FeedTheRealm-org/core-service/internal/assets-service/controllers/sprites"
	sprites_repo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/sprites"
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

	/* TODO PROTECT THESE ENDPOINTS: */
	spritesGroup.POST("/categories", spritesController.AddCategory)
	spritesGroup.PUT("", spritesController.UploadSpriteData)
}
