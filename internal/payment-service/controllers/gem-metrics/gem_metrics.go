package gem_metrics

import (
	"net/http"

	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	gem_metrics "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-metrics"
	"github.com/gin-gonic/gin"
)

type gemMetricsController struct {
	service gem_metrics.GemMetricsService
}

func NewGemMetricsController(service gem_metrics.GemMetricsService) GemMetricsController {
	return &gemMetricsController{service: service}
}

// GetMetrics godoc
// @Summary      Get gem metrics
// @Description  Returns aggregated gem metrics. Admin only.
// @Tags         payment-service
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.GemMetricsResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /payments/gems/metrics [get]
func (c *gemMetricsController) GetMetrics(ctx *gin.Context) {
	if err := common_handlers.IsAdminSession(ctx); err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	metrics, err := c.service.GetMetrics()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	flow := 0.0
	if metrics.GemsBought > 0 {
		flow = float64(metrics.GemsSpent) / float64(metrics.GemsBought)
	}

	res := dtos.GemMetricsResponse{
		GemsBought:  metrics.GemsBought,
		GemsSpent:   metrics.GemsSpent,
		GemsRevenue: metrics.GemsRevenue,
		GemsFlow:    flow,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}
