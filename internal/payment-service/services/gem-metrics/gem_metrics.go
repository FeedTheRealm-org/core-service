package gem_metrics

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	gem_metrics "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-metrics"
)

type gemMetricsService struct {
	repo gem_metrics.GemMetricsRepository
}

func NewGemMetricsService(repo gem_metrics.GemMetricsRepository) GemMetricsService {
	return &gemMetricsService{repo: repo}
}

func (s *gemMetricsService) GetMetrics() (*models.GemMetrics, error) {
	return s.repo.GetMetrics()
}
