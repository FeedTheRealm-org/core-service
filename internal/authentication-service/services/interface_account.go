package services

import (
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
)

type AccountService interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateAccount(email string, password string) (*models.User, string, error)
	LoginAccount(email string, password string) (*models.User, string, error)
	ValidateSessionToken(token string) error
	VerifyAccount(email string, code string) (bool, error)
}
