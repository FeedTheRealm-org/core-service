package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/middleware"
	_ "github.com/FeedTheRealm-org/core-service/swagger"

	"github.com/swaggo/files"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupSwaggerRouter initializes the route Group /swagger/*any and
// the corresponding handlers for the swagger documentation.
func SetupSwaggerRouter(r *gin.Engine, conf *config.Config) {
	handlers := []gin.HandlerFunc{}

	if conf.Server.Environment == config.Production {
		handlers = append(handlers, middleware.AdminCheckMiddleware())
	}
	handlers = append(handlers, ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/swagger/*any", handlers...)
}
