package gem_metrics

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gemMetricsRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewGemMetricsRepository(conf *config.Config, db *config.DB) GemMetricsRepository {
	return &gemMetricsRepository{conf: conf, db: db}
}

func (r *gemMetricsRepository) GetMetrics() (*models.GemMetrics, error) {
	metrics := models.GemMetrics{ID: 1}
	if err := r.db.Conn.Where("id = ?", 1).FirstOrCreate(&metrics).Error; err != nil {
		return nil, err
	}
	return &metrics, nil
}

func (r *gemMetricsRepository) AddGemsBoughtAndRevenue(gems int64, revenue float64) error {
	newMetrics := models.GemMetrics{
		ID:          1,
		GemsBought:  gems,
		GemsRevenue: revenue,
	}

	result := r.db.Conn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"gems_bought":  gorm.Expr("gem_metrics.gems_bought + ?", gems),
			"gems_revenue": gorm.Expr("gem_metrics.gems_revenue + ?", revenue),
			"updated_at":   gorm.Expr("NOW()"),
		}),
	}).Create(&newMetrics)

	return result.Error
}

func (r *gemMetricsRepository) AddGemsSpent(gems int64) error {
	newMetrics := models.GemMetrics{
		ID:        1,
		GemsSpent: gems,
	}

	result := r.db.Conn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"gems_spent": gorm.Expr("gem_metrics.gems_spent + ?", gems),
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(&newMetrics)

	return result.Error
}
