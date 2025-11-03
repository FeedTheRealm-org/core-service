package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	authRouter "github.com/FeedTheRealm-org/core-service/internal/authentication-service/router"
	conversionRouter "github.com/FeedTheRealm-org/core-service/internal/conversion-service/router"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	worldBrowserRouter "github.com/FeedTheRealm-org/core-service/internal/world-browser-service/router"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, conf *config.Config) {
	// Setup global middleware
	r.Use(middleware.ErrorHandlerMiddleware())

	// Setup service routers
	r.NoRoute(middleware.NotFoundController)
	authRouter.SetupAuthenticationServiceRouter(r, conf)
	conversionRouter.SetupConversionServiceRouter(r, conf)
	worldBrowserRouter.SetupWorldBrowserServiceRouter(r, conf)
	SetupSwaggerRouter(r)
}
