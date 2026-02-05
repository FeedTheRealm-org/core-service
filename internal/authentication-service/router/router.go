package router

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/controllers"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	"github.com/gin-gonic/gin"
)

func SetupAuthenticationServiceRouter(r *gin.Engine, conf *config.Config, db *config.DB, jwtManager *session.JWTManager) {
	g := r.Group("/auth")
	accountRepo, err := repositories.NewAccountRepository(conf, db)
	if err != nil {
		logger.Logger.Errorf("Failed to connect to DB: %v", err)
	}

	accountService := services.NewAccountService(conf, accountRepo, jwtManager)
	emailService := services.NewEmailSenderService(conf)
	accountController := controllers.NewAccountController(conf, accountService, emailService)

	g.POST("/signup", accountController.CreateAccount)
	g.POST("/login", accountController.LoginAccount)
	g.POST("/verify", accountController.VerifyAccount)
	g.POST("/refresh", accountController.RefreshVerification)
	g.GET("/check-session", accountController.CheckSessionExpiration)
}
