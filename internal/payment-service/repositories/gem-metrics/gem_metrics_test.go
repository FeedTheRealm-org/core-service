package gem_metrics

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/stretchr/testify/assert"
)

var gemMetricsConf *config.Config
var gemMetricsDB *config.DB
var gemMetricsRepo GemMetricsRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	gemMetricsConf = config.CreateConfig()
	var err error
	gemMetricsDB, err = config.NewDB(gemMetricsConf)
	if err != nil {
		panic(err)
	}
	gemMetricsRepo = NewGemMetricsRepository(gemMetricsConf, gemMetricsDB)

	clearGemMetricsTables()
	code := m.Run()
	clearGemMetricsTables()
	os.Exit(code)
}

func clearGemMetricsTables() {
	_ = gemMetricsDB.Conn.Exec("TRUNCATE TABLE gem_metrics RESTART IDENTITY CASCADE;")
}

func TestGemMetricsRepository_GetMetricsAndUpdates(t *testing.T) {
	clearGemMetricsTables()

	metrics, err := gemMetricsRepo.GetMetrics()
	assert.NoError(t, err)
	assert.Equal(t, int32(1), metrics.ID)
	assert.Equal(t, int64(0), metrics.GemsBought)
	assert.Equal(t, int64(0), metrics.GemsSpent)
	assert.Equal(t, float64(0), metrics.GemsRevenue)

	assert.NoError(t, gemMetricsRepo.AddGemsBoughtAndRevenue(10, 2.5))
	assert.NoError(t, gemMetricsRepo.AddGemsSpent(3))

	metrics, err = gemMetricsRepo.GetMetrics()
	assert.NoError(t, err)
	assert.Equal(t, int64(10), metrics.GemsBought)
	assert.Equal(t, int64(3), metrics.GemsSpent)
	assert.Equal(t, float64(2.5), metrics.GemsRevenue)
}
