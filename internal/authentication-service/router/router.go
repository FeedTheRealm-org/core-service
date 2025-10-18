package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/controllers"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/world-browser-service/utils/logger"
	"github.com/gin-gonic/gin"
)

func SetupAuthenticationServiceRouter(r *gin.Engine, conf *config.Config) {
	g := r.Group("/auth")
	repo, err := repositories.NewExampleRepository(conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to connect to DB: %v", err)
	}

	service := services.NewExampleService(conf, repo)
	controller := controllers.NewExampleController(conf, service)

	g.GET("/example-msg", controller.GetExample)
	g.GET("/example-query", controller.GetSumQuery)

	accountRepo, err := repositories.NewAccountRepository(conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to connect to DB: %v", err)
	}

	accountService := services.NewAccountService(conf, accountRepo)
	accountController := controllers.NewAccountController(conf, accountService)

	g.POST("/signup", accountController.CreateAccount)
}
