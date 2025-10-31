package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	authRouter "github.com/FeedTheRealm-org/core-service/internal/authentication-service/router"
	conversionRouter "github.com/FeedTheRealm-org/core-service/internal/conversion-service/router"
	worldBrowserRouter "github.com/FeedTheRealm-org/core-service/internal/world-browser-service/router"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, conf *config.Config) {
	authRouter.SetupAuthenticationServiceRouter(r, conf)
	conversionRouter.SetupConversionServiceRouter(r, conf)
	worldBrowserRouter.SetupWorldBrowserServiceRouter(r, conf)
	SetupSwaggerRouter(r)
}
