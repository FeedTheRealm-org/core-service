package controllers

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	common_dtos "github.com/FeedTheRealm-org/core-service/internal/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

type accountController struct {
	conf           *config.Config
	accountService services.AccountService
	emailService   services.EmailSenderService
}

func NewAccountController(conf *config.Config, accountService services.AccountService, emailService services.EmailSenderService) AccountController {
	return &accountController{
		conf:           conf,
		accountService: accountService,
		emailService:   emailService,
	}
}

// @Summary Sign up
// @Description Create a new user account
// @Tags authentication-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.CreateAccountRequestDTO true "Signup data"
// @Success 200  {object}  dtos.CreateAccountResponseDTO "Successful login"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /auth/signup [post]
func (ec *accountController) CreateAccount(c *gin.Context) {
	req := &dtos.CreateAccountRequestDTO{}
	if err := c.ShouldBindJSON(req); err != nil {
		logger.Logger.Errorf("CreateAccount: failed to bind JSON: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Logger.Infof("CreateAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.Logger.Info("CreateAccount: missing email")
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Email is required",
			Status:   400,
			Detail:   "You must provide an email address to create an account.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	if req.Password == "" {
		logger.Logger.Info("CreateAccount: missing password for email=%s", req.Email)
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Password is required",
			Status:   400,
			Detail:   "You must provide a password to create an account.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	result, verificationCode, err := ec.accountService.CreateAccount(req.Email, req.Password)
	if err != nil {
		if _, ok := err.(*services.AccountAlreadyExistsError); ok {
			logger.Logger.Infof("CreateAccount: account already exists for email=%s", req.Email)
			c.JSON(409, common_dtos.ErrorResponse{
				Type:     "validation",
				Title:    "Email is already in use",
				Status:   409,
				Detail:   "The email address is already in use by another account.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		if _, ok := err.(*services.AccountInvalidFormat); ok {
			logger.Logger.Infof("CreateAccount: invalid account format for email=%s: %v", req.Email, err)
			c.JSON(400, common_dtos.ErrorResponse{
				Type:     "validation",
				Title:    err.(*services.AccountInvalidFormat).Msg,
				Status:   400,
				Detail:   "The email address is not valid.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		logger.Logger.Errorf("CreateAccount: service error for email=%s: %v", req.Email, err)
		c.JSON(500, common_dtos.ErrorResponse{
			Type:     "server",
			Title:    "Internal server error",
			Status:   500,
			Detail:   "An unexpected error occurred.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	if ec.conf.Server.Environment != config.Testing {
		logger.Logger.Infof("CreateAccount: account created for email=%s", result.Email)
		err = ec.emailService.SendVerificationEmail(result.Email, verificationCode)
		if err != nil {
			logger.Logger.Errorf("CreateAccount: failed to send verification email to email=%s: %v", result.Email, err)
		}
	}

	logger.Logger.Infof("CreateAccount: account created for email=%s", result.Email)
	c.JSON(201, common_dtos.DataEnvelope[dtos.CreateAccountResponseDTO]{
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
// @Success 200  {object}  dtos.LoginAccountResponseDTO "Successful login"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request body"
// @Failure 401  {object}  dtos.ErrorResponse "Invalid credentials or invalid JWT token"
// @Router /auth/signup [post]
func (ec *accountController) LoginAccount(c *gin.Context) {
	req := dtos.LoginAccountRequestDTO{}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("LoginAccount: failed to bind JSON: %v", err)
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Invalid request body",
			Status:   400,
			Detail:   "The request body is not valid.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	logger.Logger.Infof("LoginAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.Logger.Info("LoginAccount: missing email")
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Email is required",
			Status:   400,
			Detail:   "You must provide an email address to log in.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	if req.Password == "" {
		logger.Logger.Info("LoginAccount: missing password for email=%s", req.Email)
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Password is required",
			Status:   400,
			Detail:   "You must provide a password to log in.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	token, err := ec.accountService.LoginAccount(req.Email, req.Password)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.Logger.Infof("LoginAccount: account not found for email=%s", req.Email)
			c.JSON(404, common_dtos.ErrorResponse{
				Type:     "validation",
				Title:    "Invalid email or password",
				Status:   404,
				Detail:   "The email address or password is incorrect.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		if _, ok := err.(*services.AccountNotVerifiedError); ok {
			logger.Logger.Infof("LoginAccount: account not verified for email=%s", req.Email)
			c.JSON(403, common_dtos.ErrorResponse{
				Type:     "validation",
				Title:    "Please verify your account before logging in",
				Status:   403,
				Detail:   "You must verify your email address before you can log in.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		if _, ok := err.(*services.AccountFailedToCreateTokenError); ok {
			logger.Logger.Errorf("LoginAccount: failed to create token for email=%s: %v", req.Email, err)
		}

		logger.Logger.Errorf("LoginAccount: service error for email=%s: %v", req.Email, err)
		c.JSON(500, common_dtos.ErrorResponse{
			Type:     "server",
			Title:    "Internal server error",
			Status:   500,
			Detail:   "An unexpected error occurred.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	logger.Logger.Infof("LoginAccount: login successful for email=%s", req.Email)
	c.JSON(200, common_dtos.DataEnvelope[dtos.LoginAccountResponseDTO]{
		Data: dtos.LoginAccountResponseDTO{
			Token: token,
		},
	})
}

func (ec *accountController) CheckSessionExpiration(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		logger.Logger.Info("CheckSessionExpiration: missing Authorization header")
		c.JSON(400, gin.H{"error": "Authorization header required"})
		return
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if token == "" {
		logger.Logger.Info("CheckSessionExpiration: missing token in Authorization header")
		c.JSON(400, gin.H{"error": "Token is required"})
		return
	}

	logger.Logger.Info("CheckSessionExpiration: received request to check session token from header")
	err := ec.accountService.ValidateSessionToken(token)
	if err != nil {
		if _, ok := err.(*services.AccountSessionExpired); ok {
			logger.Logger.Info("CheckSessionExpiration: session token has expired")
			c.JSON(401, gin.H{"error": "Session has expired"})
			return
		}
		logger.Logger.Errorf("CheckSessionExpiration: service error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	logger.Logger.Info("CheckSessionExpiration: session token is valid")
	c.JSON(200, gin.H{"message": "Session is valid"})
}

// @Summary Verify Account
// @Description Verify a user account with email verification code
// @Tags authentication-service
// @Accept   json
// @Produce  json
// @Param   request body dtos.VerifyAccountRequestDTO true "Verification data"
// @Success 200  {object}  dtos.VerifyAccountResponseDTO "Successful verification (Wrapped in data envelope)"
// @Failure 400  {object}  dtos.ErrorResponse "Bad request or invalid code"
// @Failure 500  {object}  dtos.ErrorResponse "Internal server error"
// @Router /auth/verify [post]
func (ec *accountController) VerifyAccount(c *gin.Context) {
	req := dtos.VerifyAccountRequestDTO{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("VerifyAccount: failed to bind JSON: %v", err)
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Invalid request body",
			Status:   400,
			Detail:   "The request body is not valid JSON.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	logger.Logger.Infof("VerifyAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.Logger.Info("VerifyAccount: missing email")
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Email is required",
			Status:   400,
			Detail:   "You must provide an email address to verify an account.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	if req.Code == "" {
		logger.Logger.Infof("VerifyAccount: missing verification code for email=%s", req.Email)
		c.JSON(400, common_dtos.ErrorResponse{
			Type:     "validation",
			Title:    "Verification code is required",
			Status:   400,
			Detail:   "You must provide a verification code.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	verified, err := ec.accountService.VerifyAccount(req.Email, req.Code)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.Logger.Infof("VerifyAccount: account not found for email=%s", req.Email)
			c.JSON(404, common_dtos.ErrorResponse{
				Type:     "not_found",
				Title:    "Account not found",
				Status:   404,
				Detail:   "No account exists with the provided email address.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		if _, ok := err.(*services.InvalidVerificationCodeError); ok {
			logger.Logger.Infof("VerifyAccount: invalid verification code for email=%s", req.Email)
			c.JSON(401, common_dtos.ErrorResponse{
				Type:     "validation",
				Title:    "Invalid verification code",
				Status:   401,
				Detail:   "The verification code is incorrect.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		if _, ok := err.(*services.VerificationCodeExpiredError); ok {
			logger.Logger.Infof("VerifyAccount: verification code expired for email=%s", req.Email)
			c.JSON(400, common_dtos.ErrorResponse{
				Type:     "validation",
				Title:    "Verification code has expired",
				Status:   400,
				Detail:   "The verification code has expired. Please request a new one.",
				Instance: c.Request.RequestURI,
			})
			return
		}

		logger.Logger.Errorf("VerifyAccount: service error for email=%s: %v", req.Email, err)
		c.JSON(500, common_dtos.ErrorResponse{
			Type:     "server",
			Title:    "Internal server error",
			Status:   500,
			Detail:   "An unexpected error occurred.",
			Instance: c.Request.RequestURI,
		})
		return
	}

	logger.Logger.Infof("VerifyAccount: account verified successfully for email=%s", req.Email)
	c.JSON(200, common_dtos.DataEnvelope[dtos.VerifyAccountResponseDTO]{
		Data: dtos.VerifyAccountResponseDTO{
			Email:    req.Email,
			Verified: verified,
		},
	})
}
