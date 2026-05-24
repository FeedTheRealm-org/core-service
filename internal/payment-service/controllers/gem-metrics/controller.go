package gem_metrics

import "github.com/gin-gonic/gin"

type GemMetricsController interface {
	// GetMetrics handles the request to retrieve gem metrics.
	GetMetrics(ctx *gin.Context)
}
