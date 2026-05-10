package creator_balances

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type creatorBalancesRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewCreatorBalancesRepository(conf *config.Config, db *config.DB) CreatorBalancesRepository {
	return &creatorBalancesRepository{conf: conf, db: db}
}

func (r *creatorBalancesRepository) AddBalance(userId uuid.UUID, amount decimal.Decimal) error {
	newBalance := models.CreatorBalance{
		UserID:  userId,
		Balance: amount,
	}

	result := r.db.Conn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"balance":    gorm.Expr("creator_balances.balance + ?", amount),
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(&newBalance)

	return result.Error
}

func (r *creatorBalancesRepository) GetBalance(userId uuid.UUID) (decimal.Decimal, error) {
	cb := models.CreatorBalance{UserID: userId, Balance: decimal.Zero}
	err := r.db.Conn.Where("user_id = ?", userId).FirstOrCreate(&cb).Error
	if err != nil {
		return decimal.Zero, err
	}
	return cb.Balance, nil
}
