package services

import (
	"math/rand"
	"strings"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	code_generator "github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/code-generator"
	validator "github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/credential-validation"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/hashing"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	"github.com/google/uuid"
)

type accountService struct {
	conf *config.Config
	repo repositories.AccountRepository
	jwt  *session.JWTManager
}

type AccountNotFoundError struct{}

func (e *AccountNotFoundError) Error() string {
	return "Account not found"
}

type AccountFailedToCreateError struct{}

func (e *AccountFailedToCreateError) Error() string {
	return "Failed to create account"
}

type AccountAlreadyExistsError struct{}

func (e *AccountAlreadyExistsError) Error() string {
	return "Account already exists"
}

type AccountFailedToCreateTokenError struct{}

func (e *AccountFailedToCreateTokenError) Error() string {
	return "Failed to create session token"
}

type AccountSessionExpired struct{}

func (e *AccountSessionExpired) Error() string {
	return "Session has expired"
}

type AccountSessionInvalid struct{}

func (e *AccountSessionInvalid) Error() string {
	return "Session is invalid"
}

type AccountNotVerifiedError struct{}

func (e *AccountNotVerifiedError) Error() string {
	return "Account not verified"
}

type InvalidVerificationCodeError struct{}

func (e *InvalidVerificationCodeError) Error() string {
	return "Invalid verification code"
}

type VerificationCodeExpiredError struct{}

func (e *VerificationCodeExpiredError) Error() string {
	return "Verification code has expired"
}

type AccountAlreadyVerifiedError struct{}

func (e *AccountAlreadyVerifiedError) Error() string {
	return "Account already verified"
}

type AccountInvalidFormat struct {
	Msg string
}

func (e *AccountInvalidFormat) Error() string {
	return "Account format is invalid"
}

func (s *accountService) seedAdminAccount(conf *config.Config) {
	adminEmail := conf.Server.AdminEmail
	adminPassword := conf.Server.AdminPassword

	if adminEmail == "" || adminPassword == "" {
		logger.Logger.Warn("Admin email or password not provided, skipping admin account creation")
		return
	}

	if _, _, _, err := s.LoginAccount(adminEmail, adminPassword, true); err == nil {
		logger.Logger.Infof("Admin account already exists with email: %s", adminEmail)
		return
	}

	_, code, err := s.CreateAccount(adminEmail, adminPassword, true)
	if err != nil {
		logger.Logger.Warnf("Failed to create admin account: %v", err)
	} else {
		logger.Logger.Infof("Admin account created with email: %s", adminEmail)
	}

	_, err = s.VerifyAccount(adminEmail, code)
	if err != nil {
		logger.Logger.Warnf("Failed to verify admin account: %v", err)
	}
}

func NewAccountService(conf *config.Config, repo repositories.AccountRepository, jwtManager *session.JWTManager) AccountService {
	newAccountService := &accountService{
		conf: conf,
		repo: repo,
		jwt:  jwtManager,
	}

	newAccountService.seedAdminAccount(conf)

	return newAccountService
}

func (s *accountService) GetUserByEmail(email string) (*models.User, error) {
	email = strings.ToLower(email)
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, &AccountNotFoundError{}
	}

	return user, nil
}

func (s *accountService) CreateAccount(email string, password string, isAdmin bool) (*models.User, string, error) {
	email = strings.ToLower(email)
	existingUser, err := s.repo.GetAccountByEmail(email)
	if err == nil && existingUser != nil {
		return nil, "", &AccountAlreadyExistsError{}
	}

	err = validator.IsValidEmail(email)
	if err != nil {
		if _, ok := err.(*validator.EmptyEmailError); ok {
			return nil, "", &AccountInvalidFormat{
				Msg: "Empty email",
			}
		}

		return nil, "", &AccountInvalidFormat{
			Msg: "Invalid email",
		}
	}

	err = validator.IsValidPassword(password)
	if err != nil {
		if _, ok := err.(*validator.EmptyPasswordError); ok {
			return nil, "", &AccountInvalidFormat{
				Msg: "Empty password",
			}
		}

		if _, ok := err.(*validator.PasswordTooShortError); ok {
			return nil, "", &AccountInvalidFormat{
				Msg: "Password is too short",
			}
		}

		if _, ok := err.(*validator.PasswordNoLetterError); ok {
			return nil, "", &AccountInvalidFormat{
				Msg: "Password must contain at least one letter",
			}
		}

		if _, ok := err.(*validator.PasswordNoNumberError); ok {
			return nil, "", &AccountInvalidFormat{
				Msg: "Password must contain at least one number",
			}
		}

		return nil, "", &AccountInvalidFormat{
			Msg: "Invalid password",
		}
	}

	hashedPassword, err := hashing.HashPassword(password)
	if err != nil {
		return nil, "", &AccountFailedToCreateError{}
	}

	functionGenerator := rand.Int
	if s.conf.Server.Environment == config.Testing {
		functionGenerator = code_generator.StaticGenerateCode
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		IsAdmin:  isAdmin,
	}

	verificationCode := code_generator.GenerateCode(functionGenerator)
	err = s.repo.CreateAccount(user, verificationCode)
	if err != nil {
		return nil, "", &AccountFailedToCreateError{}
	}

	return user, verificationCode, nil
}

func (s *accountService) LoginAccount(email string, password string, isAdminReq bool) (*models.User, string, string, error) {
	email = strings.ToLower(email)
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, "", "", &AccountNotFoundError{}
	}

	isPasswordValid := hashing.VerifyPassword(user.Password, password)
	if !isPasswordValid {
		return nil, "", "", &AccountNotFoundError{}
	}

	if !user.Verified && !isAdminReq {
		return nil, "", "", &AccountNotVerifiedError{}
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.Id.String(), user.IsAdmin)
	if err != nil {
		return nil, "", "", &AccountFailedToCreateTokenError{}
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.Id.String(), user.IsAdmin)
	if err != nil {
		return nil, "", "", &AccountFailedToCreateTokenError{}
	}

	err = s.repo.UpdateRefreshTokenUpdatedAt(user.Id, time.Now())
	if err != nil {
		return nil, "", "", &AccountFailedToCreateTokenError{}
	}

	return user, accessToken, refreshToken, nil
}

func (s *accountService) ListAccounts(query string, verified *bool, offset int, limit int) ([]models.User, int64, error) {
	return s.repo.ListAccounts(query, verified, offset, limit)
}

func (s *accountService) UpdateAdminStatus(id string, isAdmin bool) error {
	userId, err := uuid.Parse(id)
	if err != nil {
		return &AccountInvalidFormat{Msg: "invalid user id"}
	}
	return s.repo.UpdateAdminStatus(userId, isAdmin)
}

func (s *accountService) ValidateAccessToken(token string) error {
	// If a fixed server token is configured and matches the provided token, is a valid session.
	if s.conf != nil && s.conf.ServerFixedToken != "" && token == s.conf.ServerFixedToken {
		return nil
	}

	claims, err := s.jwt.IsValidateAccessToken(token, time.Now())
	if err != nil {
		if _, ok := err.(*session.JWTExpiredTokenError); ok {
			return &AccountSessionExpired{}
		}
		return &AccountSessionInvalid{}
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok || userIDStr == "" {
		return &AccountSessionInvalid{}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return &AccountSessionInvalid{}
	}

	_, err = s.repo.GetAccountById(userID)
	if err != nil {
		return &AccountSessionInvalid{}
	}

	return nil
}

func (s *accountService) ValidateRefreshToken(token string, email string) error {
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return &AccountNotFoundError{}
	}

	claims, err := s.jwt.IsValidateRefreshToken(token, time.Now(), user.RefreshTokenUpdatedAt)
	if err != nil {
		if _, ok := err.(*session.JWTExpiredTokenError); ok {
			return &AccountSessionExpired{}
		}
		return &AccountSessionInvalid{}
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok || userIDStr == "" {
		return &AccountSessionInvalid{}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return &AccountSessionInvalid{}
	}

	_, err = s.repo.GetAccountById(userID)
	if err != nil {
		return &AccountSessionInvalid{}
	}

	return nil
}

func (s *accountService) VerifyAccount(email string, code string) (bool, error) {
	email = strings.ToLower(email)
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return false, &AccountNotFoundError{}
	}

	currentTime := time.Now()
	err = s.repo.VerifyAccount(user, code, currentTime)
	if err != nil {
		if _, ok := err.(*repositories.AccountNotFoundError); ok {
			return false, &AccountNotFoundError{}
		}
		if _, ok := err.(*repositories.AccountNotVerifiedError); ok {
			return false, &InvalidVerificationCodeError{}
		}
		if _, ok := err.(*repositories.AccountVerificationExpired); ok {
			return false, &VerificationCodeExpiredError{}
		}
		return false, err
	}

	return true, nil
}

func (s *accountService) RefreshVerificationCode(email string) (string, error) {
	email = strings.ToLower(email)
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return "", &AccountNotFoundError{}
	}

	if user.Verified {
		return "", &AccountAlreadyVerifiedError{}
	}

	functionGenerator := rand.Int
	if s.conf.Server.Environment == config.Testing {
		functionGenerator = code_generator.StaticGenerateCode
	}

	newCode := code_generator.GenerateCode(functionGenerator)
	expiry := time.Now().Add(10 * time.Minute)

	if err := s.repo.RefreshVerificationCode(user, newCode, expiry); err != nil {
		return "", &AccountFailedToCreateError{}
	}

	return newCode, nil
}

func (s *accountService) RefreshToken(email string) (string, string, error) {
	email = strings.ToLower(email)
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return "", "", &AccountNotFoundError{}
	}

	if !user.Verified {
		return "", "", &AccountNotVerifiedError{}
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.Id.String(), user.IsAdmin)
	if err != nil {
		return "", "", &AccountFailedToCreateTokenError{}
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.Id.String(), user.IsAdmin)
	if err != nil {
		return "", "", &AccountFailedToCreateTokenError{}
	}

	err = s.repo.UpdateRefreshTokenUpdatedAt(user.Id, time.Now())
	if err != nil {
		return "", "", &AccountFailedToCreateTokenError{}
	}

	return accessToken, refreshToken, nil
}
