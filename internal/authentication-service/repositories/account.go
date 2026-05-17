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
			Attempts:         0,
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
	}

	if accountActivation.VerificationCode != code {
		accountActivation.Attempts += 1
		if err := ar.db.Conn.Model(&models.AccountVerification{}).Where("user_id = ?", user.Id).Update("attempts", accountActivation.Attempts).Error; err != nil {
			return &DatabaseError{message: err.Error()}
		}

		if accountActivation.Attempts >= 3 {
			if err := ar.db.Conn.Delete(&accountActivation).Error; err != nil {
				return &DatabaseError{message: err.Error()}
			}
			return &AccountNotVerifiedError{}
		}

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

func (ar *accountRepository) RefreshVerificationCode(user *models.User, verificationCode string, expiresAt time.Time) error {
	var accountActivation models.AccountVerification

	if err := ar.db.Conn.Where("user_id = ?", user.Id).First(&accountActivation).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			accountVerfication := &models.AccountVerification{
				UserId:           user.Id,
				VerificationCode: verificationCode,
				Attempts:         0,
				CreatedAt:        time.Now(),
				ExpiresAt:        expiresAt,
			}
			if err := ar.db.Conn.Create(accountVerfication).Error; err != nil {
				return &DatabaseError{message: err.Error()}
			}
			return nil
		}
		return &DatabaseError{message: err.Error()}
	}

	accountActivation.VerificationCode = verificationCode
	accountActivation.Attempts = 0
	accountActivation.ExpiresAt = expiresAt

	if err := ar.db.Conn.Model(&models.AccountVerification{}).Where("user_id = ?", user.Id).Updates(map[string]interface{}{
		"verification_code": accountActivation.VerificationCode,
		"attempts":          accountActivation.Attempts,
		"expires_at":        accountActivation.ExpiresAt,
	}).Error; err != nil {
		return &DatabaseError{message: err.Error()}
	}

	return nil
}

func (ar *accountRepository) UpdateRefreshTokenUpdatedAt(id uuid.UUID, updatedAt time.Time) error {
	if err := ar.db.Conn.Model(&models.User{}).Where("id = ?", id).Update("refresh_token_updated_at", updatedAt).Error; err != nil {
		return &DatabaseError{message: err.Error()}
	}
	return nil
}

func (ar *accountRepository) CreatePasswordReset(userID uuid.UUID, otpHash string, expiresAt time.Time) (*models.PasswordReset, error) {
	reset := &models.PasswordReset{
		UserId:       userID,
		OTPHash:      otpHash,
		OTPExpiresAt: expiresAt,
		Attempts:     0,
		OTPVerified:  false,
		Used:         false,
	}
	if err := ar.db.Conn.Create(reset).Error; err != nil {
		return nil, &DatabaseError{message: err.Error()}
	}
	return reset, nil
}

func (ar *accountRepository) GetActivePasswordResetByUserID(userID uuid.UUID) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	err := ar.db.Conn.
		Where("user_id = ? AND used = false", userID).
		Order("created_at DESC").
		First(&reset).Error
	if err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, &AccountNotFoundError{}
		}
		return nil, &DatabaseError{message: err.Error()}
	}
	return &reset, nil
}

func (ar *accountRepository) IncrementPasswordResetAttempts(resetID uuid.UUID) error {
	if err := ar.db.Conn.Model(&models.PasswordReset{}).
		Where("id = ?", resetID).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error; err != nil {
		return &DatabaseError{message: err.Error()}
	}
	return nil
}

func (ar *accountRepository) MarkPasswordResetOTPVerified(resetID uuid.UUID, resetTokenHash string, resetTokenExpiresAt time.Time) error {
	if err := ar.db.Conn.Model(&models.PasswordReset{}).
		Where("id = ?", resetID).
		Updates(map[string]interface{}{
			"otp_verified":           true,
			"reset_token_hash":       resetTokenHash,
			"reset_token_expires_at": resetTokenExpiresAt,
		}).Error; err != nil {
		return &DatabaseError{message: err.Error()}
	}
	return nil
}

func (ar *accountRepository) GetPasswordResetByTokenHash(tokenHash string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	err := ar.db.Conn.
		Where("reset_token_hash = ? AND used = false AND otp_verified = true", tokenHash).
		First(&reset).Error
	if err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, &AccountNotFoundError{}
		}
		return nil, &DatabaseError{message: err.Error()}
	}
	return &reset, nil
}

func (ar *accountRepository) InvalidateAllPasswordResets(userID uuid.UUID) error {
	if err := ar.db.Conn.Model(&models.PasswordReset{}).
		Where("user_id = ? AND used = false", userID).
		Update("used", true).Error; err != nil {
		return &DatabaseError{message: err.Error()}
	}
	return nil
}

func (ar *accountRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	if err := ar.db.Conn.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error; err != nil {
		return &DatabaseError{message: err.Error()}
	}
	return nil
}
