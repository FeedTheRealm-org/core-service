package gem_metrics

import "github.com/FeedTheRealm-org/core-service/internal/payment-service/models"

type GemMetricsService interface {
	GetMetrics() (*models.GemMetrics, error)
}
