package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-browser-service/controllers"
	"github.com/FeedTheRealm-org/core-service/internal/world-browser-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/world-browser-service/services"
	"github.com/gin-gonic/gin"
)

func SetupWorldBrowserServiceRouter(r *gin.Engine, conf *config.Config) {
	g := r.Group("/world-browser")
	repo := repositories.NewExampleRepository(conf)
	service := services.NewExampleService(conf, repo)
	controller := controllers.NewExampleController(conf, service)

	g.GET("", controller.GetExample)
}
