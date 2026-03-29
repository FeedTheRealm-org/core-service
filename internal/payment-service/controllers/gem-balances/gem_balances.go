package gem_balances

import (
	"io"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	gem_balances "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/gem-balances"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type gemBalancesController struct {
	conf              *config.Config
	gemBalanceService gem_balances.GemBalancesService
}

func NewGemBalancesController(conf *config.Config, gemBalanceService gem_balances.GemBalancesService) GemBalancesController {
	return &gemBalancesController{
		conf:              conf,
		gemBalanceService: gemBalanceService,
	}
}

// GetAllGemBalances godoc
// @Summary      List all gem balances
// @Description  Returns all gem balances. Admin only.
// @Tags         payment-service
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.GemBalanceResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
func (bc *gemBalancesController) GetAllGemBalances(c *gin.Context) {
	_, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	if err := common_handlers.IsAdminSession(c); err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	balances, err := bc.gemBalanceService.GetAllGemBalances()
	if err != nil {
		_ = c.Error(err)
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, balances)
}

// GetGemBalanceByUserId godoc
// @Summary      Get current user gem balance
// @Description  Returns the authenticated user's gem balance. Creates one if missing.
// @Tags         payment-service
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.GemBalanceResponse
// @Failure      401  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /payments/gems/balances [get]
func (bc *gemBalancesController) GetGemBalanceByUserId(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	balance, err := bc.gemBalanceService.GetGemBalanceByUserId(userId)
	if balance == nil {
		err = bc.gemBalanceService.CreateGemBalance(userId)
		if err != nil {
			_ = c.Error(err)
			return
		}
		balance, err = bc.gemBalanceService.GetGemBalanceByUserId(userId)
		if err != nil {
			_ = c.Error(err)
			return
		}
	} else if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.GemBalanceResponse{
		UserId: balance.UserId,
		Gems:   balance.Gems,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}

// UpdateGemBalance godoc
// @Summary      Update a user's gem balance
// @Description  Updates gem balance for a user. Admin only.
// @Tags         payment-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "User ID"
// @Param        request  body      dtos.UpdateGemBalanceRequest  true  "Gem balance update payload"
// @Success      200      {object}  dtos.GemBalanceResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /payments/gems/balances/{id} [put]
func (bc *gemBalancesController) UpdateGemBalance(c *gin.Context) {
	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid user_id: " + err.Error()))
		return
	}

	req := dtos.UpdateGemBalanceRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	if err := bc.gemBalanceService.UpdateGemBalance(userId, req.Gems); err != nil {
		_ = c.Error(err)
		return
	}

	balance, err := bc.gemBalanceService.GetGemBalanceByUserId(userId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.GemBalanceResponse{
		UserId: balance.UserId,
		Gems:   balance.Gems,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}

// CreateCheckoutSession godoc
// @Summary      Create checkout session
// @Description  Creates a checkout session URL to purchase a gem pack.
// @Tags         payment-service
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.CheckoutRequest   true  "Checkout request"
// @Success      200      {object}  dtos.CheckoutResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /payments/checkout [post]
func (bc *gemBalancesController) CreateCheckoutSession(c *gin.Context) {
	userId, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	req := dtos.CheckoutRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	checkoutUrl, err := bc.gemBalanceService.CreateCheckoutSession(userId, req.GemPackId, req.SuccessUrl, req.CancelUrl)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &dtos.CheckoutResponse{
		CheckoutUrl: checkoutUrl,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}

// HandleStripeWebhook godoc
// @Summary      Handle Stripe webhook
// @Description  Processes Stripe webhook events for completed payments.
// @Tags         payment-service
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.WebhookResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /payments/webhook/stripe [post]
func (bc *gemBalancesController) HandleStripeWebhook(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(65536))

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		_ = c.Error(errors.NewBadRequestError("failed to read request body"))
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		_ = c.Error(errors.NewBadRequestError("missing Stripe-Signature header"))
		return
	}

	if err := bc.gemBalanceService.HandleWebhook(body, signature); err != nil {
		_ = c.Error(errors.NewInternalServerError("cannot process webhook: " + err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, &dtos.WebhookResponse{})
}
