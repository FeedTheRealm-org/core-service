package gem_metrics

import (
	"errors"
	"testing"

	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/stretchr/testify/assert"
)

type fakeGemMetricsRepo struct {
	metrics *models.GemMetrics
	err     error
}

func (f *fakeGemMetricsRepo) GetMetrics() (*models.GemMetrics, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.metrics, nil
}

func (f *fakeGemMetricsRepo) AddGemsBoughtAndRevenue(gems int64, revenue float64) error {
	return nil
}

func (f *fakeGemMetricsRepo) AddGemsSpent(gems int64) error {
	return nil
}

func TestGemMetricsService_GetMetrics(t *testing.T) {
	repo := &fakeGemMetricsRepo{metrics: &models.GemMetrics{ID: 1, GemsBought: 10}}
	svc := NewGemMetricsService(repo)

	metrics, err := svc.GetMetrics()
	assert.NoError(t, err)
	assert.Equal(t, int32(1), metrics.ID)
	assert.Equal(t, int64(10), metrics.GemsBought)
}

func TestGemMetricsService_GetMetrics_Error(t *testing.T) {
	repo := &fakeGemMetricsRepo{err: errors.New("boom")}
	svc := NewGemMetricsService(repo)

	_, err := svc.GetMetrics()
	assert.Error(t, err)
}
