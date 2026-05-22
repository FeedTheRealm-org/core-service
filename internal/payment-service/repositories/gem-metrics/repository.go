package gem_metrics

import "github.com/FeedTheRealm-org/core-service/internal/payment-service/models"

type GemMetricsRepository interface {
	GetMetrics() (*models.GemMetrics, error)
	AddGemsBoughtAndRevenue(gems int64, revenue float64) error
	AddGemsSpent(gems int64) error
}
