package services

import "github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"

type AccountService interface {
	GetUserByEmail(email string) (*repositories.User, error)
	CreateAccount(email string, password string) (*repositories.User, error)
	LoginAccount(email string, password string) (string, error)
	ValidateSessionToken(token string) error
}
