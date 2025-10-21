package services

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/hashing"
	jwt "github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/session-token"
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
