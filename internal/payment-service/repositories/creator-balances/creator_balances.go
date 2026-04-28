package creator_balances

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
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

func (r *creatorBalancesRepository) AddBalance(userId uuid.UUID, amount float64) error {
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

func (r *creatorBalancesRepository) GetBalance(userId uuid.UUID) (float64, error) {
	var cb models.CreatorBalance
	err := r.db.Conn.Where("user_id = ?", userId).First(&cb).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return cb.Balance, nil
}
