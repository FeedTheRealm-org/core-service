package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/controllers"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/gin-gonic/gin"
)

func SetupAuthenticationServiceRouter(r *gin.Engine, conf *config.Config) {
	g := r.Group("/auth")
	repo := repositories.NewExampleRepository(conf)
	service := services.NewExampleService(conf, repo)
	controller := controllers.NewExampleController(conf, service)

	g.GET("", controller.GetExample)
}
