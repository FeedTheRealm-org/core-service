package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	gem_balances_controller "github.com/FeedTheRealm-org/core-service/internal/payment-service/controllers/gem-balances"
	gem_packs_controller "github.com/FeedTheRealm-org/core-service/internal/payment-service/controllers/gem-packs"
	gem_balances_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-balances"
	gem_packs_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	gem_balances_service "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-balances"
	gem_packs_service "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-packs"
	"github.com/gin-gonic/gin"
)

func SetupGemPacksServiceRouter(conf *config.Config, db *config.DB, g *gin.RouterGroup) {
	packsRepo := gem_packs_repo.NewGemPacksRepository(conf, db)
	gemGemPacksService := gem_packs_service.NewGemPacksService(conf, packsRepo)
	gemGemPacksController := gem_packs_controller.NewGemPacksController(conf, gemGemPacksService)

	/* Packs Endpoints */
	packsGroup := g.Group("/packs")
	packsGroup.GET("", gemGemPacksController.GetAllGemPacks)
	packsGroup.GET("/:id", gemGemPacksController.GetGemPackById)
	packsGroup.POST("", middleware.AdminCheckMiddleware(), gemGemPacksController.CreateGemPack)
	packsGroup.PUT("/:id", middleware.AdminCheckMiddleware(), gemGemPacksController.UpdateGemPack)
	packsGroup.DELETE("/:id", middleware.AdminCheckMiddleware(), gemGemPacksController.DeleteGemPack)
}

func SetupBalancesServiceRouter(conf *config.Config, db *config.DB, paymentGroup *gin.RouterGroup, gemsGroup *gin.RouterGroup) {
	gemBalancesRepo := gem_balances_repo.NewGemBalancesRepository(conf, db)
	packsRepo := gem_packs_repo.NewGemPacksRepository(conf, db)
	gemBalancesService := gem_balances_service.NewGemBalancesService(conf, gemBalancesRepo, packsRepo)
	gemBalancesController := gem_balances_controller.NewGemBalancesController(conf, gemBalancesService)

	/* Balances Endpoints */
	balancesGroup := gemsGroup.Group("/balances")
	// balancesGroup.GET("", gemBalancesController.GetAllGemBalances)
	balancesGroup.GET("", gemBalancesController.GetGemBalanceByUserId)
	balancesGroup.PUT("/:id", middleware.AdminCheckMiddleware(), gemBalancesController.UpdateGemBalance)

	/* Webhook Endpoint */
	paymentGroup.POST("/checkout", gemBalancesController.CreateCheckoutSession)
	paymentGroup.POST("/webhook/stripe", gemBalancesController.HandleStripeWebhook)
}

func SetupPaymentServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	paymentGroup := r.Group("/payments")
	gemsGroup := paymentGroup.Group("/gems")

	SetupGemPacksServiceRouter(conf, db, gemsGroup)
	SetupBalancesServiceRouter(conf, db, paymentGroup, gemsGroup)

	return nil
}
