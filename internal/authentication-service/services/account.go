package services

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	validator "github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/credential-validation"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/hashing"
	jwt "github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/session"
)

type accountService struct {
	conf *config.Config
	repo repositories.AccountRepository
	jwt  *jwt.JWTManager
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

type AccountInvalidFormat struct{
	Msg string
}

func (e *AccountInvalidFormat) Error() string {
	return "Account format is invalid"
}

func NewAccountService(conf *config.Config, repo repositories.AccountRepository) AccountService {
	return &accountService{
		conf: conf,
		repo: repo,
		jwt:  jwt.NewJWTManager(conf.SessionTokenSecretKey, conf.SessionTokenDuration),
	}
}

func (s *accountService) GetUserByEmail(email string) (*repositories.User, error) {
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return nil, &AccountNotFoundError{}
	}

	return user, nil
}

func (s *accountService) CreateAccount(email string, password string) (*repositories.User, error) {
	existingUser, err := s.repo.GetAccountByEmail(email)
	if err == nil && existingUser != nil {
		return nil, &AccountAlreadyExistsError{}
	}

	err = validator.IsValidEmail(email)
	if err != nil {
		if _, ok := err.(*validator.EmptyEmailError); ok {
			return nil, &AccountInvalidFormat{
				Msg: "Empty email",
			}
		}

		return nil, &AccountInvalidFormat{
			Msg: "Invalid email",
		}
	}

	err = validator.IsValidPassword(password)
	if err != nil {
		if _, ok := err.(*validator.EmptyPasswordError); ok {
			return nil, &AccountInvalidFormat{
				Msg: "Empty password",
			}
		}

		if _, ok := err.(*validator.PasswordTooShortError); ok {
			return nil, &AccountInvalidFormat{
				Msg: "Password is too short",
			}
		}

		if _, ok := err.(*validator.PasswordNoLetterError); ok {
			return nil, &AccountInvalidFormat{
				Msg: "Password must contain at least one letter",
			}
		}

		if _, ok := err.(*validator.PasswordNoNumberError); ok {
			return nil, &AccountInvalidFormat{
				Msg: "Password must contain at least one number",
			}
		}

		return nil, &AccountInvalidFormat{
			Msg: "Invalid password",
		}
	}

	hashedPassword, err := hashing.HashPassword(password)
	if err != nil {
		return nil, &AccountFailedToCreateError{}
	}

	user := &repositories.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	err = s.repo.CreateAccount(user)
	if err != nil {
		return nil, &AccountFailedToCreateError{}
	}

	return user, nil
}

func (s *accountService) LoginAccount(email string, password string) (string, error) {
	user, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return "", &AccountNotFoundError{}
	}

	isPasswordValid := hashing.VerifyPassword(user.PasswordHash, password)
	if !isPasswordValid {
		return "", &AccountNotFoundError{}
	}

	token, err := s.jwt.GenerateToken(user.Email)
	if err != nil {
		return "", &AccountFailedToCreateTokenError{}
	}

	return token, nil
}

func (s *accountService) ValidateSessionToken(token string) error {
	if err := s.jwt.IsValidateToken(token, time.Now()); err != nil {
		if _, ok := err.(*jwt.JWTExpiredTokenError); ok {
			return &AccountSessionExpired{}
		}
		return &AccountSessionInvalid{}
	}

	return nil
}
