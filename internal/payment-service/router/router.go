package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	creator_balances_controller "github.com/FeedTheRealm-org/core-service/internal/payment-service/controllers/creator-balances"
	gem_balances_controller "github.com/FeedTheRealm-org/core-service/internal/payment-service/controllers/gem-balances"
	gem_packs_controller "github.com/FeedTheRealm-org/core-service/internal/payment-service/controllers/gem-packs"
	zones_subscriptions_controller "github.com/FeedTheRealm-org/core-service/internal/payment-service/controllers/zones-subscriptions"
	creator_balances_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/creator-balances"
	gem_balances_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-balances"
	gem_packs_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	zones_subscriptions_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/zones-subscriptions"
	creator_balances_service "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/creator-balances"
	gem_balances_service "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-balances"
	gem_packs_service "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-packs"
	zones_subscriptions_service "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/zones-subscriptions"
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
	creatorBalanceRepo := creator_balances_repo.NewCreatorBalancesRepository(conf, db)
	packsRepo := gem_packs_repo.NewGemPacksRepository(conf, db)

	gemBalancesService := gem_balances_service.NewGemBalancesService(conf, gemBalancesRepo, packsRepo, creatorBalanceRepo)

	gemBalancesController := gem_balances_controller.NewGemBalancesController(conf, gemBalancesService)

	/* Balances Endpoints */
	balancesGroup := gemsGroup.Group("/balances")
	balancesGroup.GET("", gemBalancesController.GetGemBalanceByUserId)
	balancesGroup.PUT("/:id", middleware.AdminCheckMiddleware(), gemBalancesController.UpdateGemBalance)

	/* Purchase Endpoints */
	paymentGroup.POST("/purchase/:cosmetic_id", gemBalancesController.PurchaseCosmetic)

	/* Webhook Endpoint for Gems */
	paymentGroup.POST("/checkout", gemBalancesController.CreateCheckoutSession)
	paymentGroup.POST("/webhook/stripe", gemBalancesController.HandleStripeWebhook)
}

func SetupSubscriptionsServiceRouter(conf *config.Config, db *config.DB, subscriptionGroup *gin.RouterGroup) {
	zonesSubscriptionsRepo := zones_subscriptions_repo.NewSubscriptionRepository(conf, db)
	zonesSubscriptionsService := zones_subscriptions_service.NewSubscriptionService(conf, zonesSubscriptionsRepo)
	zonesSubscriptionsController := zones_subscriptions_controller.NewZonesSubscriptionsController(conf, zonesSubscriptionsService)

	// External / user-facing subscription routes
	subscriptionGroup.POST("/checkout", zonesSubscriptionsController.CreateCheckoutSession)
	subscriptionGroup.PUT("/slots", zonesSubscriptionsController.UpdateSlots)
	subscriptionGroup.GET("/status", zonesSubscriptionsController.GetStatus)
	subscriptionGroup.GET("/pricing", zonesSubscriptionsController.GetPricingInfo)
	subscriptionGroup.DELETE("", zonesSubscriptionsController.CancelSubscription)

	// Webhook for Subscriptions
	subscriptionGroup.POST("/webhook/stripe", zonesSubscriptionsController.HandleWebhook)

	// Internal routes bypassed by JWT
	internalGroup := subscriptionGroup.Group("/internal")
	internalGroup.GET("/users/:user_id/status", zonesSubscriptionsController.CheckInternalAvailability)
	internalGroup.PUT("/users/:user_id/used-slots", zonesSubscriptionsController.InternalUpdateUsedSlots)
}

func SetupCreatorBalancesRouter(conf *config.Config, db *config.DB, paymentGroup *gin.RouterGroup) {
	creatorBalancesRepo := creator_balances_repo.NewCreatorBalancesRepository(conf, db)
	creatorBalancesService := creator_balances_service.NewCreatorBalancesService(creatorBalancesRepo)
	creatorBalancesController := creator_balances_controller.NewCreatorBalancesController(creatorBalancesService)

	/* Creator Balances Endpoints */
	paymentGroup.GET("/balances/creators/:creator_id", creatorBalancesController.GetBalance)
}

func SetupPaymentServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	paymentGroup := r.Group("/payments")
	subscriptionGroup := r.Group("/subscriptions")
	gemsGroup := paymentGroup.Group("/gems")

	SetupGemPacksServiceRouter(conf, db, gemsGroup)
	SetupBalancesServiceRouter(conf, db, paymentGroup, gemsGroup)
	SetupSubscriptionsServiceRouter(conf, db, subscriptionGroup)
	SetupCreatorBalancesRouter(conf, db, paymentGroup)

	return nil
}
