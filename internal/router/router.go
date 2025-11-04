package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	authRouter "github.com/FeedTheRealm-org/core-service/internal/authentication-service/router"
	conversionRouter "github.com/FeedTheRealm-org/core-service/internal/conversion-service/router"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	worldBrowserRouter "github.com/FeedTheRealm-org/core-service/internal/world-browser-service/router"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, conf *config.Config, db *config.DB) {
	jwtManager := session.NewJWTManager(conf.SessionTokenSecretKey, conf.SessionTokenDuration)

	// Setup global middleware
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(middleware.JWTAuthMiddleware(jwtManager))

	// Setup service routers
	r.NoRoute(middleware.NotFoundController)
	authRouter.SetupAuthenticationServiceRouter(r, conf, db, jwtManager)
	conversionRouter.SetupConversionServiceRouter(r, conf)
	worldBrowserRouter.SetupWorldBrowserServiceRouter(r, conf)
	SetupSwaggerRouter(r)
}
