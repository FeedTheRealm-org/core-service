package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assetsRouter "github.com/FeedTheRealm-org/core-service/internal/assets-service/router"
	authRouter "github.com/FeedTheRealm-org/core-service/internal/authentication-service/router"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	itemsRouter "github.com/FeedTheRealm-org/core-service/internal/items-service/router"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	playersRouter "github.com/FeedTheRealm-org/core-service/internal/players-service/router"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	worldRouter "github.com/FeedTheRealm-org/core-service/internal/world-service/router"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, conf *config.Config, db *config.DB) error {
	jwtManager := session.NewJWTManager(conf.SessionTokenSecretKey, conf.SessionTokenDuration)

	// Setup global middleware
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(middleware.JWTAuthMiddleware(jwtManager, conf.ServerFixedToken))

	// Setup service routers
	r.NoRoute(common_handlers.NotFoundController)

	if err := authRouter.SetupAuthenticationServiceRouter(r, conf, db, jwtManager); err != nil {
		return err
	}

	if err := playersRouter.SetupPlayerServiceRouter(r, conf, db); err != nil {
		return err
	}

	if err := worldRouter.SetupWorldServiceRouter(r, conf, db); err != nil {
		return err
	}

	if err := assetsRouter.SetupAssetsServiceRouter(r, conf, db); err != nil {
		return err
	}

	if err := itemsRouter.SetupItemsServiceRouter(r, conf, db); err != nil {
		return err
	}

	if conf.Server.Environment != config.Production {
		SetupSwaggerRouter(r)
	}

	return nil
}
