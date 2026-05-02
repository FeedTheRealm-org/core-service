package gem_balances

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DatabaseError struct {
	message string
}

func (e *DatabaseError) Error() string {
	return "Database error occurred: " + e.message
}

type gemBalancesRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewGemBalancesRepository(conf *config.Config, db *config.DB) GemBalancesRepository {
	return &gemBalancesRepository{
		conf: conf,
		db:   db,
	}
}

func (br *gemBalancesRepository) CreateGemBalance(userId uuid.UUID) error {
	balance := models.GemBalance{
		UserId: userId,
		Gems:   0,
	}

	return br.db.Conn.Create(&balance).Error
}

func (br *gemBalancesRepository) GetAllGemBalances() ([]*models.GemBalance, error) {
	var balances []*models.GemBalance

	if err := br.db.Conn.Find(&balances).Error; err != nil {
		return nil, &DatabaseError{message: err.Error()}
	}

	return balances, nil
}

func (br *gemBalancesRepository) GetGemBalanceByUserId(userId uuid.UUID) (*models.GemBalance, error) {
	var balance models.GemBalance

	if err := br.db.Conn.Where("user_id = ?", userId).First(&balance).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, errors.NewNotFoundError("balance not found")
		}
		return nil, &DatabaseError{message: err.Error()}
	}

	return &balance, nil
}

func (br *gemBalancesRepository) AddToGemBalance(userId uuid.UUID, gems int64) error {
	return br.db.Conn.Model(&models.GemBalance{}).Where("user_id = ?", userId).UpdateColumn("gems", gorm.Expr("gems + ?", gems)).Error
}

func (br *gemBalancesRepository) ApplyStripeCheckoutCreditIfUnprocessed(userId uuid.UUID, gems int64, eventID string, sessionID string) (bool, error) {
	applied := false

	err := br.db.Conn.Transaction(func(tx *gorm.DB) error {
		processedEvent := &models.ProcessedStripeWebhookEvent{
			EventID:   eventID,
			SessionID: sessionID,
			EventType: "checkout.session.completed",
			UserID:    userId,
		}

		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(processedEvent)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return nil
		}

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.GemBalance{UserId: userId, Gems: 0}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.GemBalance{}).Where("user_id = ?", userId).UpdateColumn("gems", gorm.Expr("gems + ?", gems)).Error; err != nil {
			return err
		}

		applied = true
		return nil
	})

	if err != nil {
		return false, &DatabaseError{message: err.Error()}
	}

	return applied, nil
}

func (br *gemBalancesRepository) UpdateGemBalance(userId uuid.UUID, newGems int64) error {
	return br.db.Conn.Model(&models.GemBalance{}).Where("user_id = ?", userId).Update("gems", newGems).Error
}
