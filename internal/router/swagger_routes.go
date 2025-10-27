package router

import (
	_ "github.com/FeedTheRealm-org/core-service/docs"

	"github.com/swaggo/files"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupSwaggerRouter initializes the route Group /swagger/*any and
// the corresponding handlers for the swagger documentation.
func SetupSwaggerRouter(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
