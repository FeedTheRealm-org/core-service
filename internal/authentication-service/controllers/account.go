package controllers

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

type accountController struct {
	conf    *config.Config
	service services.AccountService
}

func NewAccountController(conf *config.Config, service services.AccountService) AccountController {
	return &accountController{
		conf:    conf,
		service: service,
	}
}

func (ec *accountController) CreateAccount(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("CreateAccount: failed to bind JSON: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.GetLogger().Infof("CreateAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.GetLogger().Info("CreateAccount: missing email")
		c.JSON(400, gin.H{"error": "Email is required"})
		return
	}

	if req.Password == "" {
		logger.GetLogger().Info("CreateAccount: missing password for email=%s", req.Email)
		c.JSON(400, gin.H{"error": "Password is required"})
		return
	}

	result, err := ec.service.CreateAccount(req.Email, req.Password)
	if err != nil {
		if _, ok := err.(*services.AccountAlreadyExistsError); ok {
			logger.GetLogger().Infof("CreateAccount: account already exists for email=%s", req.Email)
			c.JSON(400, gin.H{"error": "Email is already in use"})
			return
		}
		logger.GetLogger().Errorf("CreateAccount: service error for email=%s: %v", req.Email, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	logger.GetLogger().Infof("CreateAccount: account created for email=%s", result.Email)
	c.JSON(201, gin.H{"message": "Account created successfully", "email": result.Email})
}
