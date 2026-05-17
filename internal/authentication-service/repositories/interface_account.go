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

	// Password reset
	CreatePasswordReset(userID uuid.UUID, otpHash string, expiresAt time.Time) (*models.PasswordReset, error)
	GetActivePasswordResetByUserID(userID uuid.UUID) (*models.PasswordReset, error)
	IncrementPasswordResetAttempts(resetID uuid.UUID) error
	MarkPasswordResetOTPVerified(resetID uuid.UUID, resetTokenHash string, resetTokenExpiresAt time.Time) error
	GetPasswordResetByTokenHash(tokenHash string) (*models.PasswordReset, error)
	InvalidateAllPasswordResets(userID uuid.UUID) error
	UpdatePassword(userID uuid.UUID, hashedPassword string) error
}
