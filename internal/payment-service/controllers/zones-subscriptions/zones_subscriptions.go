package zones_subscriptions

import (
	"io"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	zones_subscriptions "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/zones-subscriptions"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

type subscriptionController struct {
	zonesSubscriptionsService zones_subscriptions.SubscriptionService
}

func NewZonesSubscriptionsController(conf *config.Config, zonesSubscriptionsService zones_subscriptions.SubscriptionService) ZonesSubscriptionsController {
	return &subscriptionController{zonesSubscriptionsService: zonesSubscriptionsService}
}

func (zc *subscriptionController) CreateCheckoutSession(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	var req dtos.CheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	url, err := zc.zonesSubscriptionsService.CreateCheckoutSession(userID, req.Slots, req.SuccessUrl, req.CancelUrl)
	if err != nil {
		logger.Logger.Error("Failed to create checkout session for user " + userID.String() + ": " + err.Error())
		_ = c.Error(errors.NewInternalServerError("Failed to create checkout session for user: " + err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, &dtos.CheckoutResponse{CheckoutUrl: url})
}

func (zc *subscriptionController) UpdateSlots(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		logger.Logger.Error("Failed to parse user_id from context: " + err.Error())
		_ = c.Error(errors.NewBadRequestError("Invalid user_id in context"))
		return
	}

	var req dtos.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	sub, err := zc.zonesSubscriptionsService.UpdateSlots(userID, req.Slots)
	if err != nil {
		logger.Logger.Error("Failed to update subscription slots for user " + userID.String() + ": " + err.Error())
		_ = c.Error(errors.NewInternalServerError("Failed to update subscription slots: " + err.Error()))
		return
	}

	res := &dtos.SubscriptionStatusResponse{
		Slots:           sub.TotalSlots,
		UsedSlots:       sub.UsedSlots,
		Status:          sub.Status,
		NextBillingDate: sub.NextBillingDate,
		AmountDue:       sub.AmountDue,
	}
	common_handlers.HandleSuccessResponse(c, 200, res)
}

func (zc *subscriptionController) CancelSubscription(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		logger.Logger.Error("Failed to parse user_id from context: " + err.Error())
		_ = c.Error(errors.NewBadRequestError("Invalid user_id in context"))
		return
	}

	sub, err := zc.zonesSubscriptionsService.CancelSubscription(userID)
	if err != nil {
		logger.Logger.Error("Failed to cancel subscription for user " + userID.String() + ": " + err.Error())
		_ = c.Error(errors.NewInternalServerError("Failed to cancel subscription: " + err.Error()))
		return
	}

	res := &dtos.SubscriptionStatusResponse{
		Slots:           sub.TotalSlots,
		UsedSlots:       sub.UsedSlots,
		Status:          sub.Status,
		NextBillingDate: sub.NextBillingDate,
		AmountDue:       sub.AmountDue,
	}
	common_handlers.HandleSuccessResponse(c, 200, res)
}

func (zc *subscriptionController) GetStatus(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(errors.NewUnauthorizedError(err.Error()))
		return
	}

	sub, err := zc.zonesSubscriptionsService.GetByUserID(userID)
	if err != nil {
		logger.Logger.Warn("User " + userID.String() + " does not have an active subscription")
		_ = c.Error(errors.NewNotFoundError("User does not have an active subscription"))
		return
	}

	res := &dtos.SubscriptionStatusResponse{
		Slots:           sub.TotalSlots,
		UsedSlots:       sub.UsedSlots,
		Status:          sub.Status,
		NextBillingDate: sub.NextBillingDate,
		AmountDue:       sub.AmountDue,
	}
	common_handlers.HandleSuccessResponse(c, 200, res)
}

func (zc *subscriptionController) CheckInternalAvailability(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		logger.Logger.Error("Failed to parse user_id from context: " + err.Error())
		_ = c.Error(errors.NewBadRequestError("Invalid user_id in context"))
		return
	}

	var req dtos.InternalSlotsCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	allowed, totalSlots, err := zc.zonesSubscriptionsService.CheckAvalibility(userID, req.RequiredSlots)
	if err != nil {
		logger.Logger.Warn("User " + userID.String() + " does not have an active subscription")
		_ = c.Error(errors.NewUnauthorizedError("User does not have an active subscription"))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, dtos.InternalSlotsCheckResponse{
		Allowed:    allowed,
		TotalSlots: totalSlots,
	})
}

func (zc *subscriptionController) HandleWebhook(c *gin.Context) {
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

	if err := zc.zonesSubscriptionsService.HandleWebhook(body, signature); err != nil {
		_ = c.Error(errors.NewInternalServerError("cannot process webhook: " + err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, &dtos.WebhookResponse{})
}
