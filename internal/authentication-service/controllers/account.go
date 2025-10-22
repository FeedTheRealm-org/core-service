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
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	logger.GetLogger().Infof("CreateAccount: account created for email=%s", result.Email)
	c.JSON(201, gin.H{"message": "Account created successfully", "email": result.Email})
}

func (ec *accountController) LoginAccount(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("LoginAccount: failed to bind JSON: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.GetLogger().Infof("LoginAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.GetLogger().Info("LoginAccount: missing email")
		c.JSON(400, gin.H{"error": "Email is required"})
		return
	}

	if req.Password == "" {
		logger.GetLogger().Info("LoginAccount: missing password for email=%s", req.Email)
		c.JSON(400, gin.H{"error": "Password is required"})
		return
	}

	token, err := ec.service.LoginAccount(req.Email, req.Password)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.GetLogger().Infof("LoginAccount: account not found for email=%s", req.Email)
			c.JSON(404, gin.H{"error": "Invalid email or password"})
			return
		}

		if _, ok := err.(*services.AccountFailedToCreateTokenError); ok {
			logger.GetLogger().Errorf("LoginAccount: failed to create token for email=%s: %v", req.Email, err)
		}

		logger.GetLogger().Errorf("LoginAccount: service error for email=%s: %v", req.Email, err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	logger.GetLogger().Infof("LoginAccount: login successful for email=%s", req.Email)
	c.JSON(200, gin.H{"message": "Login successful", "token": token})
}

func (ec *accountController) CheckSessionExpiration(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		logger.GetLogger().Info("CheckSessionExpiration: missing Authorization header")
		c.JSON(400, gin.H{"error": "Authorization header required"})
		return
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if token == "" {
		logger.GetLogger().Info("CheckSessionExpiration: missing token in Authorization header")
		c.JSON(400, gin.H{"error": "Token is required"})
		return
	}

	logger.GetLogger().Info("CheckSessionExpiration: received request to check session token from header")
	err := ec.service.ValidateSessionToken(token)
	if err != nil {
		if _, ok := err.(*services.AccountSessionExpired); ok {
			logger.GetLogger().Info("CheckSessionExpiration: session token has expired")
			c.JSON(401, gin.H{"error": "Session has expired"})
			return
		}
		logger.GetLogger().Errorf("CheckSessionExpiration: service error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	logger.GetLogger().Info("CheckSessionExpiration: session token is valid")
	c.JSON(200, gin.H{"message": "Session is valid"})
}
