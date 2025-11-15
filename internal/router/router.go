package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	authRouter "github.com/FeedTheRealm-org/core-service/internal/authentication-service/router"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	playersRouter "github.com/FeedTheRealm-org/core-service/internal/players-service/router"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	worldRouter "github.com/FeedTheRealm-org/core-service/internal/world-service/router"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, conf *config.Config, db *config.DB) {
	jwtManager := session.NewJWTManager(conf.SessionTokenSecretKey, conf.SessionTokenDuration)

	// Setup global middleware
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(middleware.JWTAuthMiddleware(jwtManager))

	// Setup service routers
	r.NoRoute(common_handlers.NotFoundController)

	authRouter.SetupAuthenticationServiceRouter(r, conf, db, jwtManager)
	playersRouter.SetupPlayerServiceRouter(r, conf, db)
	worldRouter.SetupWorldServiceRouter(r, conf, db)

	if conf.Server.Environment != config.Production {
		SetupSwaggerRouter(r)
	}
}
