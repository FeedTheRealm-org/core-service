package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	packs_controller "github.com/FeedTheRealm-org/core-service/internal/payment-service/controllers/packs"
	packs_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/packs"
	packs_service "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/packs"
	"github.com/gin-gonic/gin"
)

func SetupPacksServiceRouter(conf *config.Config, db *config.DB, g *gin.RouterGroup) {
	packsRepo := packs_repo.NewPacksRepository(conf, db)
	packsService := packs_service.NewPacksService(conf, packsRepo)
	packsController := packs_controller.NewPacksController(conf, packsService)

	/* Packs Endpoints */
	packsGroup := g.Group("/packs")
	packsGroup.GET("", packsController.GetAllPacks)
	packsGroup.GET("/:id", packsController.GetPackById)
	packsGroup.POST("", packsController.CreatePack)
	packsGroup.PUT("/:id", packsController.UpdatePack)
	packsGroup.DELETE("/:id", packsController.DeletePack)
}

func SetupPaymentServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	g := r.Group("/payments")

	SetupPacksServiceRouter(conf, db, g)

	return nil
}
