package creator_balances

import (
	"net/http"

	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	creator_balances "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/creator-balances"
	"github.com/gin-gonic/gin"
)

type creatorBalancesController struct {
	service creator_balances.CreatorBalancesService
}

func NewCreatorBalancesController(service creator_balances.CreatorBalancesService) CreatorBalancesController {
	return &creatorBalancesController{service: service}
}

// GetBalance godoc
// @Summary      Get creator balance
// @Description  Get the current creator balance of the authenticated user
// @Tags         payment-service
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.CreatorBalanceResponse
// @Failure      401  {object} dtos.ErrorResponse
// @Failure      500  {object} dtos.ErrorResponse
// @Router       /payments/balances/creators [get]
func (c *creatorBalancesController) GetBalance(ctx *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(ctx)
	if err != nil {
		_ = ctx.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	balance, err := c.service.GetBalance(userId)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	balanceParsed, _ := balance.Float64()

	res := dtos.CreatorBalanceResponse{
		UserID:  userId,
		Balance: balanceParsed,
	}

	common_handlers.HandleSuccessResponse(ctx, http.StatusOK, res)
}
