package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/conversion-service/controllers"
	"github.com/FeedTheRealm-org/core-service/internal/conversion-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/conversion-service/services"
	"github.com/gin-gonic/gin"
)

func SetupConversionServiceRouter(r *gin.Engine, conf *config.Config) {
	g := r.Group("/conversion")
	repo := repositories.NewExampleRepository(conf)
	service := services.NewExampleService(conf, repo)
	controller := controllers.NewExampleController(conf, service)

	g.GET("", controller.GetExample)
}
