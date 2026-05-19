package repositories

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"github.com/google/uuid"
)

type AccountRepository interface {
	GetAccountById(id uuid.UUID) (*models.User, error)
	GetAccountByEmail(email string) (*models.User, error)
	CreateAccount(user *models.User, verificationCode string) error
	VerifyAccount(user *models.User, code string, currentTime time.Time) error
	RefreshVerificationCode(user *models.User, verificationCode string, expiresAt time.Time) error
	UpdateRefreshTokenUpdatedAt(id uuid.UUID, updatedAt time.Time) error
	ListAccounts(query string, verified *bool, offset int, limit int) ([]models.User, int64, error)
	UpdateAdminStatus(id uuid.UUID, isAdmin bool) error
}
