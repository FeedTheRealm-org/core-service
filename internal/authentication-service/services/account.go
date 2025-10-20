package services

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/hashing"
)

type accountService struct {
	conf *config.Config
	repo repositories.AccountRepository
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

func NewAccountService(conf *config.Config, repo repositories.AccountRepository) AccountService {
	return &accountService{
		conf: conf,
		repo: repo,
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
