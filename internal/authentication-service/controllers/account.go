package controllers

import (
	"net/http"
	"strconv"

	"github.com/FeedTheRealm-org/core-service/config"
	dtos "github.com/FeedTheRealm-org/core-service/internal/authentication-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
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

// @Summary      Create a new account
// @Description  Creates a new user account with the provided email and password.
// @Tags         authentication-service
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateAccountRequestDTO true "Account creation details"
// @Success      201  {object}  dtos.CreateAccountResponseDTO
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      409  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /auth/signup [post]
func (ec *accountController) CreateAccount(c *gin.Context) {
	req := &dtos.CreateAccountRequestDTO{}
	if err := c.ShouldBindJSON(req); err != nil {
		logger.Logger.Errorf("CreateAccount: failed to bind JSON: %v", err)
		_ = c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	logger.Logger.Infof("CreateAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.Logger.Info("CreateAccount: missing email")
		_ = c.Error(errors.NewBadRequestError("You must provide an email address to create an account."))
		return
	}

	if req.Password == "" {
		logger.Logger.Info("CreateAccount: missing password for email=%s", req.Email)
		_ = c.Error(errors.NewBadRequestError("You must provide a password to create an account."))
		return
	}

	if req.IsAdmin == nil {
		isAdmin := false
		req.IsAdmin = &isAdmin
	}

	if err := common_handlers.IsAdminSession(c); err != nil && req.IsAdmin != nil && *req.IsAdmin {
		logger.Logger.Info("CreateAccount: Non-admin account try to create admin account with email=%s", req.Email)
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	result, verificationCode, err := ec.accountService.CreateAccount(req.Email, req.Password, *req.IsAdmin)
	if err != nil {
		if _, ok := err.(*services.AccountAlreadyExistsError); ok {
			logger.Logger.Infof("CreateAccount: account already exists for email=%s", req.Email)
			_ = c.Error(errors.NewConflictError("The email address is already in use by another account."))
			return
		}

		if _, ok := err.(*services.AccountInvalidFormat); ok {
			logger.Logger.Infof("CreateAccount: invalid account format for email=%s: %v", req.Email, err)
			_ = c.Error(errors.NewBadRequestError("The email address is not valid."))
			return
		}

		logger.Logger.Errorf("CreateAccount: service error for email=%s: %v", req.Email, err)
		_ = c.Error(errors.NewInternalServerError("An unexpected error occurred."))
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
	res := &dtos.CreateAccountResponseDTO{
		Email: result.Email,
	}

	common_handlers.HandleSuccessResponse(c, http.StatusCreated, res)
}

// @Summary      Login account
// @Description  Authenticates a user and returns an access token.
// @Tags         authentication-service
// @Accept       json
// @Produce      json
// @Param        request body dtos.LoginAccountRequestDTO true "Login credentials"
// @Success      200  {object}  dtos.LoginAccountResponseDTO
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      403  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /auth/login [post]
func (ec *accountController) LoginAccount(c *gin.Context) {
	req := dtos.LoginAccountRequestDTO{}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("LoginAccount: failed to bind JSON: %v", err)
		_ = c.Error(errors.NewBadRequestError("The request body is not valid."))
		return
	}

	logger.Logger.Infof("LoginAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.Logger.Info("LoginAccount: missing email")
		_ = c.Error(errors.NewBadRequestError("You must provide an email address to log in."))
		return
	}

	if req.Password == "" {
		logger.Logger.Info("LoginAccount: missing password for email=%s", req.Email)
		_ = c.Error(errors.NewBadRequestError("You must provide a password to log in."))
		return
	}

	isAdminReq := false
	if err := common_handlers.IsAdminSession(c); err == nil {
		isAdminReq = true
	}

	user, accessToken, refreshToken, err := ec.accountService.LoginAccount(req.Email, req.Password, isAdminReq)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.Logger.Infof("LoginAccount: account not found for email=%s", req.Email)
			_ = c.Error(errors.NewUnauthorizedError("The email address or password is incorrect."))
			return
		}

		if _, ok := err.(*services.AccountNotVerifiedError); ok {
			logger.Logger.Infof("LoginAccount: account not verified for email=%s", req.Email)
			_ = c.Error(errors.NewForbiddenError("You must verify your email address before you can log in."))
			return
		}

		if _, ok := err.(*services.AccountFailedToCreateTokenError); ok {
			logger.Logger.Errorf("LoginAccount: failed to create token for email=%s: %v", req.Email, err)
		}

		logger.Logger.Errorf("LoginAccount: service error for email=%s: %v", req.Email, err)
		_ = c.Error(errors.NewInternalServerError("An unexpected error occurred."))
		return
	}

	logger.Logger.Infof("LoginAccount: login successful for email=%s", req.Email)

	res := &dtos.LoginAccountResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Id:           user.Id.String(),
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary      Check session expiration
// @Description  Validates the integrity and expiration of an active token.
// @Tags         authentication-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.CheckSessionResponseDTO
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /auth/check-session [get]
func (ec *accountController) CheckSessionExpiration(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		logger.Logger.Info("CheckSessionExpiration: missing Authorization header")
		_ = c.Error(errors.NewBadRequestError("Authorization header required"))
		return
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if token == "" {
		logger.Logger.Info("CheckSessionExpiration: missing token in Authorization header")
		_ = c.Error(errors.NewUnauthorizedError("Token is required"))
		return
	}

	logger.Logger.Info("CheckSessionExpiration: received request to check session token from header")
	err := ec.accountService.ValidateAccessToken(token)
	if err != nil {
		if _, ok := err.(*services.AccountSessionExpired); ok {
			logger.Logger.Info("CheckSessionExpiration: session token has expired")
			_ = c.Error(errors.NewUnauthorizedError("Session has expired"))
			return
		}
		logger.Logger.Errorf("CheckSessionExpiration: service error: %v", err)
		_ = c.Error(errors.NewUnauthorizedError("Invalid session token"))
		return
	}

	logger.Logger.Info("CheckSessionExpiration: session token is valid")

	res := &dtos.CheckSessionResponseDTO{
		Message: "Session is valid",
	}
	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary      Verify an account
// @Description  Verifies a user's email address with a provided verification code.
// @Tags         authentication-service
// @Accept       json
// @Produce      json
// @Param        request body dtos.VerifyAccountRequestDTO true "Verification details"
// @Success      200  {object}  dtos.VerifyAccountResponseDTO
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /auth/verify [post]
func (ec *accountController) VerifyAccount(c *gin.Context) {
	req := dtos.VerifyAccountRequestDTO{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("VerifyAccount: failed to bind JSON: %v", err)
		_ = c.Error(errors.NewBadRequestError("The request body is not valid JSON."))
		return
	}

	logger.Logger.Infof("VerifyAccount: received request for email=%s", req.Email)

	if req.Email == "" {
		logger.Logger.Info("VerifyAccount: missing email")
		_ = c.Error(errors.NewBadRequestError("You must provide an email address to verify an account."))
		return
	}

	if req.Code == "" {
		logger.Logger.Infof("VerifyAccount: missing verification code for email=%s", req.Email)
		_ = c.Error(errors.NewBadRequestError("You must provide a verification code."))
		return
	}

	verified, err := ec.accountService.VerifyAccount(req.Email, req.Code)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.Logger.Infof("VerifyAccount: account not found for email=%s", req.Email)
			_ = c.Error(errors.NewNotFoundError("No account exists with the provided email address."))
			return
		}

		if _, ok := err.(*services.InvalidVerificationCodeError); ok {
			logger.Logger.Infof("VerifyAccount: invalid verification code for email=%s", req.Email)
			_ = c.Error(errors.NewUnauthorizedError("The verification code is incorrect."))
			return
		}

		if _, ok := err.(*services.VerificationCodeExpiredError); ok {
			logger.Logger.Infof("VerifyAccount: verification code expired for email=%s", req.Email)
			_ = c.Error(errors.NewBadRequestError("The verification code has expired. Please request a new one."))
			return
		}

		logger.Logger.Errorf("VerifyAccount: service error for email=%s: %v", req.Email, err)
		_ = c.Error(errors.NewInternalServerError("An unexpected error occurred."))
		return
	}

	logger.Logger.Infof("VerifyAccount: account verified successfully for email=%s", req.Email)
	res := &dtos.VerifyAccountResponseDTO{
		Email:    req.Email,
		Verified: verified,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary      Refresh verification code
// @Description  Generates and sends a new verification code to the user's email.
// @Tags         authentication-service
// @Accept       json
// @Produce      json
// @Param        request body dtos.RefreshVerificationRequestDTO true "Email details"
// @Success      200  {object}  dtos.RefreshVerificationResponseDTO
// @Failure      400  {object} dtos.ErrorResponse
// @Failure      404  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /auth/refresh [post]
func (ec *accountController) RefreshVerification(c *gin.Context) {
	req := dtos.RefreshVerificationRequestDTO{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("RefreshVerification: failed to bind JSON: %v", err)
		_ = c.Error(errors.NewBadRequestError("The request body is not valid JSON."))
		return
	}

	if req.Email == "" {
		logger.Logger.Info("RefreshVerification: missing email")
		_ = c.Error(errors.NewBadRequestError("You must provide an email address to refresh the verification code."))
		return
	}

	newCode, err := ec.accountService.RefreshVerificationCode(req.Email)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.Logger.Infof("RefreshVerification: account not found for email=%s", req.Email)
			_ = c.Error(errors.NewNotFoundError("No account exists with the provided email address."))
			return
		}

		if _, ok := err.(*services.AccountAlreadyVerifiedError); ok {
			logger.Logger.Infof("RefreshVerification: account already verified for email=%s", req.Email)
			_ = c.Error(errors.NewBadRequestError("The account is already verified; no verification code will be generated."))
			return
		}

		logger.Logger.Errorf("RefreshVerification: service error for email=%s: %v", req.Email, err)
		_ = c.Error(errors.NewInternalServerError("An unexpected error occurred."))
		return
	}

	if err := ec.emailService.SendVerificationEmail(req.Email, newCode); err != nil {
		logger.Logger.Errorf("RefreshVerification: failed to send verification email to email=%s: %v", req.Email, err)
	}

	logger.Logger.Infof("RefreshVerification: verification code refreshed for email=%s", req.Email)
	res := &dtos.RefreshVerificationResponseDTO{
		Email: req.Email,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary      Refresh access token
// @Description  Validates the provided refresh token and issues a new access token and refresh token pair.
// @Tags         authentication-service
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                        true  "Bearer <refresh_token>"
// @Param        request        body      dtos.RefreshTokenRequestDTO   true  "Email details"
// @Success      200            {object}  dtos.RefreshTokenResponseDTO
// @Failure      400            {object}  dtos.ErrorResponse
// @Failure      401            {object}  dtos.ErrorResponse
// @Failure      404            {object}  dtos.ErrorResponse
// @Failure      500            {object}  dtos.ErrorResponse
// @Router       /auth/refresh-token [post]
func (ec *accountController) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		logger.Logger.Info("RefreshToken: missing Authorization header")
		_ = c.Error(errors.NewBadRequestError("Authorization header required"))
		return
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if token == "" {
		logger.Logger.Info("RefreshToken: missing token in Authorization header")
		_ = c.Error(errors.NewBadRequestError("Token is required"))
		return
	}

	req := dtos.RefreshTokenRequestDTO{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Errorf("RefreshToken: failed to bind JSON: %v", err)
		_ = c.Error(errors.NewBadRequestError("The request body is not valid JSON."))
		return
	}

	logger.Logger.Infof("RefreshToken: received token refresh request for email=%s", req.Email)

	if req.Email == "" {
		logger.Logger.Info("RefreshToken: missing email")
		_ = c.Error(errors.NewBadRequestError("You must provide an email address to refresh the token."))
		return
	}

	logger.Logger.Info("RefreshToken: received request to check session token from header")
	err := ec.accountService.ValidateRefreshToken(token, req.Email)
	if err != nil {
		if _, ok := err.(*services.AccountSessionExpired); ok {
			logger.Logger.Info("RefreshToken: session token has expired")
			_ = c.Error(errors.NewUnauthorizedError("Session has expired"))
			return
		}
		logger.Logger.Errorf("RefreshToken: service error: %v", err)
		_ = c.Error(errors.NewUnauthorizedError("Invalid session token"))
		return
	}

	newAccessToken, newRefreshToken, err := ec.accountService.RefreshToken(req.Email)
	if err != nil {
		if _, ok := err.(*services.AccountNotFoundError); ok {
			logger.Logger.Infof("RefreshToken: account not found for email=%s", req.Email)
			_ = c.Error(errors.NewNotFoundError("No account exists with the provided email address."))
			return
		}

		logger.Logger.Errorf("RefreshToken: service error for email=%s: %v", req.Email, err)
		_ = c.Error(errors.NewInternalServerError("An unexpected error occurred."))
		return
	}

	logger.Logger.Infof("RefreshToken: token refreshed successfully for email=%s", req.Email)
	res := &dtos.RefreshTokenResponseDTO{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	common_handlers.HandleSuccessResponse(c, http.StatusOK, res)
}

// @Summary      List users
// @Description  Returns a paginated list of users. Admin only.
// @Tags         authentication-service
// @Security     BearerAuth
// @Produce      json
// @Param        query    query     string  false  "Filter by email"
// @Param        verified query     bool    false  "Filter by verification"
// @Param        offset   query     int     false  "Offset" default(0)
// @Param        limit    query     int     false  "Limit" default(50)
// @Success      200      {object}  dtos.UsersListResponseDTO
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /auth/users [get]
func (ec *accountController) ListUsers(c *gin.Context) {
	if err := common_handlers.IsAdminSession(c); err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	query := c.Query("query")
	verifiedParam := c.Query("verified")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		_ = c.Error(errors.NewBadRequestError("invalid offset"))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		_ = c.Error(errors.NewBadRequestError("invalid limit"))
		return
	}

	var verified *bool
	if verifiedParam != "" {
		parsed, err := strconv.ParseBool(verifiedParam)
		if err != nil {
			_ = c.Error(errors.NewBadRequestError("invalid verified filter"))
			return
		}
		verified = &parsed
	}

	users, total, err := ec.accountService.ListAccounts(query, verified, offset, limit)
	if err != nil {
		_ = c.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	response := dtos.UsersListResponseDTO{
		Users:      make([]dtos.UserSummaryResponseDTO, len(users)),
		TotalCount: total,
	}
	for i, user := range users {
		response.Users[i] = dtos.UserSummaryResponseDTO{
			ID:        user.Id.String(),
			Email:     user.Email,
			Verified:  user.Verified,
			IsAdmin:   user.IsAdmin,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	common_handlers.HandleSuccessResponse(c, http.StatusOK, response)
}

// @Summary      Update admin status
// @Description  Updates the admin flag for a user. Admin only.
// @Tags         authentication-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string                          true  "User UUID"
// @Param        request body      dtos.UpdateAdminStatusRequestDTO true  "Admin status"
// @Success      204      {string}  string "No Content"
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /auth/users/{id}/admin [put]
func (ec *accountController) UpdateAdminStatus(c *gin.Context) {
	if err := common_handlers.IsAdminSession(c); err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	req := dtos.UpdateAdminStatusRequestDTO{}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body"))
		return
	}

	if err := ec.accountService.UpdateAdminStatus(c.Param("id"), req.IsAdmin); err != nil {
		_ = c.Error(errors.NewInternalServerError(err.Error()))
		return
	}

	common_handlers.HandleBodilessResponse(c, http.StatusNoContent)
}
