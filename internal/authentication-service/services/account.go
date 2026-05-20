package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	mathrand "math/rand"
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

const (
	passwordResetOTPExpiry   = 10 * time.Minute
	passwordResetTokenExpiry = 15 * time.Minute
	passwordResetMaxAttempts = 5
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

type PasswordResetNotFoundError struct{}

func (e *PasswordResetNotFoundError) Error() string {
	return "Password reset request not found or expired"
}

type PasswordResetExpiredError struct{}

func (e *PasswordResetExpiredError) Error() string {
	return "Password reset code has expired"
}

type PasswordResetMaxAttemptsError struct{}

func (e *PasswordResetMaxAttemptsError) Error() string {
	return "Too many incorrect attempts"
}

type InvalidPasswordResetCodeError struct{}

func (e *InvalidPasswordResetCodeError) Error() string {
	return "Invalid password reset code"
}

type PasswordResetTokenExpiredError struct{}

func (e *PasswordResetTokenExpiredError) Error() string {
	return "Password reset token has expired"
}

type PasswordResetTokenAlreadyUsedError struct{}

func (e *PasswordResetTokenAlreadyUsedError) Error() string {
	return "Password reset token has already been used"
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

	functionGenerator := mathrand.Int
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

	functionGenerator := mathrand.Int
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

// generateOTP returns a cryptographically secure 6-digit numeric OTP and its bcrypt hash.
func generateOTP() (plaintext string, hash string, err error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", "", err
	}
	plaintext = fmt.Sprintf("%06d", n.Int64()+100000)
	hash, err = hashing.HashPassword(plaintext)
	if err != nil {
		return "", "", err
	}
	return plaintext, hash, nil
}

// generateResetToken returns a cryptographically secure hex token and its SHA-256 hex hash for DB lookup.
func generateResetToken() (plaintext string, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	plaintext = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plaintext))
	tokenHash = hex.EncodeToString(sum[:])
	return plaintext, tokenHash, nil
}

// ForgotPassword generates a password reset OTP for the given email.
// It always returns success to the caller to avoid leaking account existence.
// Returns the plaintext OTP (for email delivery) and nil error when a reset was created.
// Returns ("", nil) when the email is unknown or rate-limited — caller must treat all cases as success.
func (s *accountService) ForgotPassword(email string) (string, error) {
	email = strings.ToLower(email)

	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		logger.Logger.Infof("ForgotPassword: account not found for email=%s, silently skipping", email)
		return "", nil
	}

	otpPlain, otpHash, err := generateOTP()
	if err != nil {
		logger.Logger.Errorf("ForgotPassword: failed to generate OTP for user=%s: %v", user.Id, err)
		return "", &AccountFailedToCreateError{}
	}

	// Invalidate any pending resets before creating a new one.
	if err := s.repo.InvalidateAllPasswordResets(user.Id); err != nil {
		logger.Logger.Warnf("ForgotPassword: failed to invalidate old resets for user=%s: %v", user.Id, err)
	}

	expiresAt := time.Now().Add(passwordResetOTPExpiry)
	if _, err := s.repo.CreatePasswordReset(user.Id, otpHash, expiresAt); err != nil {
		logger.Logger.Errorf("ForgotPassword: failed to create password reset record for user=%s: %v", user.Id, err)
		return "", &AccountFailedToCreateError{}
	}

	logger.Logger.Infof("ForgotPassword: password reset OTP created for user=%s", user.Id)
	return otpPlain, nil
}

// VerifyPasswordResetCode validates the OTP and, on success, issues a single-use reset token.
func (s *accountService) VerifyPasswordResetCode(email string, code string) (string, error) {
	email = strings.ToLower(email)

	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return "", &PasswordResetNotFoundError{}
	}

	reset, err := s.repo.GetActivePasswordResetByUserID(user.Id)
	if err != nil {
		return "", &PasswordResetNotFoundError{}
	}

	if reset.OTPVerified {
		return "", &PasswordResetNotFoundError{}
	}

	if time.Now().After(reset.OTPExpiresAt) {
		return "", &PasswordResetExpiredError{}
	}

	if reset.Attempts >= passwordResetMaxAttempts {
		return "", &PasswordResetMaxAttemptsError{}
	}

	if !hashing.VerifyPassword(reset.OTPHash, code) {
		if err := s.repo.IncrementPasswordResetAttempts(reset.Id); err != nil {
			logger.Logger.Errorf("VerifyPasswordResetCode: failed to increment attempts for reset=%s: %v", reset.Id, err)
			return "", &PasswordResetMaxAttemptsError{}
		}
		remaining := passwordResetMaxAttempts - (reset.Attempts + 1)
		if remaining <= 0 {
			return "", &PasswordResetMaxAttemptsError{}
		}
		return "", &InvalidPasswordResetCodeError{}
	}

	tokenPlain, tokenHash, err := generateResetToken()
	if err != nil {
		logger.Logger.Errorf("VerifyPasswordResetCode: failed to generate reset token for user=%s: %v", user.Id, err)
		return "", &AccountFailedToCreateTokenError{}
	}

	resetTokenExpiresAt := time.Now().Add(passwordResetTokenExpiry)
	if err := s.repo.MarkPasswordResetOTPVerified(reset.Id, tokenHash, resetTokenExpiresAt); err != nil {
		logger.Logger.Errorf("VerifyPasswordResetCode: failed to mark OTP verified for reset=%s: %v", reset.Id, err)
		return "", &AccountFailedToCreateTokenError{}
	}

	logger.Logger.Infof("VerifyPasswordResetCode: OTP verified, reset token issued for user=%s", user.Id)
	return tokenPlain, nil
}

// ResetPassword validates the reset token, updates the password, and revokes all active sessions.
func (s *accountService) ResetPassword(resetToken string, newPassword string) error {
	sum := sha256.Sum256([]byte(resetToken))
	tokenHash := hex.EncodeToString(sum[:])

	reset, err := s.repo.GetPasswordResetByTokenHash(tokenHash)
	if err != nil {
		return &PasswordResetNotFoundError{}
	}

	if reset.ResetTokenExpiresAt == nil || time.Now().After(*reset.ResetTokenExpiresAt) {
		return &PasswordResetTokenExpiredError{}
	}

	if err := validator.IsValidPassword(newPassword); err != nil {
		if _, ok := err.(*validator.EmptyPasswordError); ok {
			return &AccountInvalidFormat{Msg: "Empty password"}
		}
		if _, ok := err.(*validator.PasswordTooShortError); ok {
			return &AccountInvalidFormat{Msg: "Password is too short"}
		}
		if _, ok := err.(*validator.PasswordNoLetterError); ok {
			return &AccountInvalidFormat{Msg: "Password must contain at least one letter"}
		}
		if _, ok := err.(*validator.PasswordNoNumberError); ok {
			return &AccountInvalidFormat{Msg: "Password must contain at least one number"}
		}
		return &AccountInvalidFormat{Msg: "Invalid password"}
	}

	hashedPassword, err := hashing.HashPassword(newPassword)
	if err != nil {
		return &AccountFailedToCreateError{}
	}

	if err := s.repo.UpdatePassword(reset.UserId, hashedPassword); err != nil {
		logger.Logger.Errorf("ResetPassword: failed to update password for user=%s: %v", reset.UserId, err)
		return &AccountFailedToCreateError{}
	}

	// Invalidate all reset records for this user (including the current one).
	if err := s.repo.InvalidateAllPasswordResets(reset.UserId); err != nil {
		logger.Logger.Warnf("ResetPassword: failed to invalidate password resets for user=%s: %v", reset.UserId, err)
	}

	// Rotate refresh_token_updated_at to force relogin on all devices.
	if err := s.repo.UpdateRefreshTokenUpdatedAt(reset.UserId, time.Now()); err != nil {
		logger.Logger.Errorf("ResetPassword: failed to revoke sessions for user=%s: %v", reset.UserId, err)
	}

	logger.Logger.Infof("ResetPassword: password reset completed for user=%s", reset.UserId)
	return nil
}
