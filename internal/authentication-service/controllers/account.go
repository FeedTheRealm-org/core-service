package controllers

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/dtos"
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

// @Summary Sign up
// @Description Create a new user account
// @Tags authentication-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.CreateAccountRequestDTO true "Signup data"
// @Success 200  {object}  dtos.CreateAccountResponseDTO "Successful login (Wrapped in data envelope)"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /auth/signup [post]
func (ec *accountController) CreateAccount(c *gin.Context) {
	req := &dtos.CreateAccountRequestDTO{}
	if err := c.ShouldBindJSON(req); err != nil {
		logger.GetLogger().Errorf("CreateAccount: failed to bind JSON: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.GetLogger().Infof("CreateAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.GetLogger().Info("CreateAccount: missing email")
		c.JSON(400, dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Email is required",
			Status:   400,
			Detail:   "You must provide an email address to create an account.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	if req.Password == "" {
		logger.GetLogger().Info("CreateAccount: missing password for email=%s", req.Email)
		c.JSON(400, dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Password is required",
			Status:   400,
			Detail:   "You must provide a password to create an account.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	result, err := ec.service.CreateAccount(req.Email, req.Password)
	if err != nil {
		if _, ok := err.(*services.AccountAlreadyExistsError); ok {
			logger.GetLogger().Infof("CreateAccount: account already exists for email=%s", req.Email)
			c.JSON(400, dtos.ErrorResponse{
				Type:     "validation",
				Title:    "Email is already in use",
				Status:   400,
				Detail:   "The email address is already in use by another account.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		if _, ok := err.(*services.AccountInvalidFormat); ok {
			logger.GetLogger().Infof("CreateAccount: invalid account format for email=%s: %v", req.Email, err)
			c.JSON(400, dtos.ErrorResponse{
				Type:     "validation",
				Title:    err.(*services.AccountInvalidFormat).Msg,
				Status:   400,
				Detail:   "The email address is not valid.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		logger.GetLogger().Errorf("CreateAccount: service error for email=%s: %v", req.Email, err)
		c.JSON(500, dtos.ErrorResponse{
			Type:     "server",
			Title:    "Internal server error",
			Status:   500,
			Detail:   "An unexpected error occurred.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	logger.GetLogger().Infof("CreateAccount: account created for email=%s", result.Email)
	c.JSON(201, dtos.DataEnvelope[dtos.CreateAccountResponseDTO]{
		Data: dtos.CreateAccountResponseDTO{
			Email: result.Email,
		},
	})

}

// @Summary Login
// @Description Log in an existing user
// @Tags authentication-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.LoginAccountRequestDTO true "Login data"
// @Success 200  {object}  dtos.LoginAccountResponseDTO "Successful login (Wrapped in data envelope)"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /auth/signup [post]
func (ec *accountController) LoginAccount(c *gin.Context) {
	req := dtos.LoginAccountRequestDTO{}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("LoginAccount: failed to bind JSON: %v", err)
		c.JSON(400, dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Invalid request body",
			Status:   400,
			Detail:   "The request body is not valid.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	logger.GetLogger().Infof("LoginAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.GetLogger().Info("LoginAccount: missing email")
		c.JSON(400, dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Email is required",
			Status:   400,
			Detail:   "You must provide an email address to log in.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	if req.Password == "" {
		logger.GetLogger().Info("LoginAccount: missing password for email=%s", req.Email)
		c.JSON(400, dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Password is required",
			Status:   400,
			Detail:   "You must provide a password to log in.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	token, err := ec.service.LoginAccount(req.Email, req.Password)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.GetLogger().Infof("LoginAccount: account not found for email=%s", req.Email)
			c.JSON(404, dtos.ErrorResponse{
				Type:     "validation",
				Title:    "Invalid email or password",
				Status:   404,
				Detail:   "The email address or password is incorrect.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		if _, ok := err.(*services.AccountFailedToCreateTokenError); ok {
			logger.GetLogger().Errorf("LoginAccount: failed to create token for email=%s: %v", req.Email, err)
		}

		logger.GetLogger().Errorf("LoginAccount: service error for email=%s: %v", req.Email, err)
		c.JSON(500, dtos.ErrorResponse{
			Type:     "server",
			Title:    "Internal server error",
			Status:   500,
			Detail:   "An unexpected error occurred.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	logger.GetLogger().Infof("LoginAccount: login successful for email=%s", req.Email)
	c.JSON(200, dtos.DataEnvelope[dtos.LoginAccountResponseDTO]{
		Data: dtos.LoginAccountResponseDTO{
			Token: token,
		},
	})
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
