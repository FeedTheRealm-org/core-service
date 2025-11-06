package repositories

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountNotFoundError struct{}

type AccountNotVerifiedError struct{}

type AccountVerificationExpired struct{}

type DatabaseError struct {
	message string
}

func (e *AccountNotFoundError) Error() string {
	return "Account not found"
}

func (e *AccountNotVerifiedError) Error() string {
	return "Account not verified"
}

func (e *AccountVerificationExpired) Error() string {
	return "Account verification has expired"
}

func (e *DatabaseError) Error() string {
	return "Database error occurred: " + e.message
}

type accountRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewAccountRepository(conf *config.Config, db *config.DB) (AccountRepository, error) {
	return &accountRepository{
		conf: conf,
		db:   db,
	}, nil
}

func (ar *accountRepository) GetAccountById(id uuid.UUID) (*models.User, error) {
	var user models.User

	if err := ar.db.Conn.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, errors.NewNotFoundError("user not found")
		}
		return nil, &DatabaseError{message: err.Error()}
	}

	return &user, nil
}

func (ar *accountRepository) GetAccountByEmail(email string) (*models.User, error) {
	var user models.User

	if err := ar.db.Conn.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, errors.NewNotFoundError("user not found")
		}
		return nil, &DatabaseError{message: err.Error()}
	}

	return &user, nil
}

func (ar *accountRepository) CreateAccount(user *models.User, verificationCode string) error {
	user.Verified = false // Set user as unverified by default
	return ar.db.Conn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return &DatabaseError{message: err.Error()}
		}

		accountVerfication := &models.AccountVerification{
			UserId:           user.Id,
			VerificationCode: verificationCode,
		}
		if err := tx.Create(accountVerfication).Error; err != nil {
			return &DatabaseError{message: err.Error()}
		}

		return nil
	})
}

func (ar *accountRepository) VerifyAccount(user *models.User, code string, currentTime time.Time) error {
	var accountActivation models.AccountVerification
	if err := ar.db.Conn.Where("user_id = ?", user.Id).First(&accountActivation).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return &AccountNotFoundError{}
		}
		return &DatabaseError{message: err.Error()}
	}

	if accountActivation.ExpiresAt.Before(currentTime) {
		return &AccountVerificationExpired{}
	} else if accountActivation.VerificationCode != code {
		return &AccountNotVerifiedError{}
	}

	err := ar.db.Conn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", user.Id).Update("verified", true).Error; err != nil {
			return &DatabaseError{message: err.Error()}
		}

		if err := tx.Delete(&accountActivation).Error; err != nil {
			return &DatabaseError{message: err.Error()}
		}

		return nil
	})
	if err != nil {
		return err
	}

	logger.Logger.Debugf("Current time: %v, Verification expiry time: %v", currentTime, accountActivation.ExpiresAt)

	user.Verified = true

	return nil
}
