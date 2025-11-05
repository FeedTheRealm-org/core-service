package repositories

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"github.com/google/uuid"
)

type User struct {
	Email        string
	PasswordHash string
	VerifyCode   string
	Expiration   time.Time
}

type AccountRepository interface {
	GetAccountByEmail(email string) (*models.User, error)
	CreateAccount(user *models.User, verificationCode string) error
	VerifyAccount(id uuid.UUID, code string, currentTime time.Time) error
}
