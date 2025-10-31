package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/controllers"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

func SetupAuthenticationServiceRouter(r *gin.Engine, conf *config.Config) {
	g := r.Group("/auth")
	accountRepo, err := repositories.NewAccountRepository(conf)
	if err != nil {
		logger.GetLogger().Errorf("Failed to connect to DB: %v", err)
	}

	accountService := services.NewAccountService(conf, accountRepo)
	accountController := controllers.NewAccountController(conf, accountService)

	g.POST("/signup", accountController.CreateAccount)
	g.POST("/login", accountController.LoginAccount)
	g.GET("/check-session", accountController.CheckSessionExpiration)
}
