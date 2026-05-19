package zones_subscriptions

import (
	"fmt"
	"io"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/common_handlers"
	custom_errors "github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/dtos"
	zones_subscriptions "github.com/FeedTheRealm-org/core-service/internal/payment-service/services/zones-subscriptions"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type subscriptionController struct {
	zonesSubscriptionsService zones_subscriptions.SubscriptionService
}

func NewZonesSubscriptionsController(conf *config.Config, zonesSubscriptionsService zones_subscriptions.SubscriptionService) ZonesSubscriptionsController {
	return &subscriptionController{zonesSubscriptionsService: zonesSubscriptionsService}
}

// CreateCheckoutSession godoc
// @Summary      Create subscription checkout session
// @Description  Creates a Stripe checkout session for a zone subscription.
// @Tags         payment-service-subscriptions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.CheckoutSessionRequest  true  "Checkout session request payload"
// @Success      200      {object}  dtos.CheckoutResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /subscriptions/checkout [post]
func (zc *subscriptionController) CreateCheckoutSession(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(custom_errors.NewUnauthorizedError(err.Error()))
		return
	}

	var req dtos.CheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(custom_errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	url, err := zc.zonesSubscriptionsService.CreateCheckoutSession(userID, req.Slots, req.SuccessUrl, req.CancelUrl)
	if err != nil {
		logger.Logger.Error("Failed to create checkout session for user " + userID.String() + ": " + err.Error())
		_ = c.Error(custom_errors.NewInternalServerError("Failed to create checkout session for user: " + err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, &dtos.CheckoutResponse{CheckoutUrl: url})
}

// UpdateSlots godoc
// @Summary      Update subscription slots
// @Description  Updates the total number of slots in the active subscription.
// @Tags         payment-service-subscriptions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.UpdateSubscriptionRequest  true  "Update subscription slots payload"
// @Success      200      {object}  dtos.SubscriptionStatusResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /subscriptions/slots [put]
func (zc *subscriptionController) UpdateSlots(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		logger.Logger.Error("Failed to parse user_id from context: " + err.Error())
		_ = c.Error(custom_errors.NewBadRequestError("Invalid user_id in context"))
		return
	}

	var req dtos.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(custom_errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	sub, err := zc.zonesSubscriptionsService.UpdateSlots(userID, req.Slots)
	if err != nil {
		logger.Logger.Info(err.Error())
		if _, ok := err.(*zones_subscriptions.CannotExceedTotalSlotsError); ok {
			logger.Logger.Warnf("User %s attempted to reduce slots to %d, but is currently using more slots than that", userID.String(), req.Slots)
			_ = c.Error(custom_errors.NewBadRequestError(fmt.Sprintf("cannot reduce slots to %d because you are currently using more than that. Please delete some zones first", req.Slots)))
			return
		}
		logger.Logger.Error("Failed to update subscription slots for user " + userID.String() + ": " + err.Error())
		_ = c.Error(custom_errors.NewInternalServerError("Failed to update subscription slots: " + err.Error()))
		return
	}

	amountDue, _ := sub.AmountDue.Float64()

	res := &dtos.SubscriptionStatusResponse{
		Slots:           sub.TotalSlots,
		UsedSlots:       sub.UsedSlots,
		Status:          string(sub.Status),
		NextBillingDate: sub.NextBillingDate,
		AmountDue:       amountDue,
	}
	common_handlers.HandleSuccessResponse(c, 200, res)
}

// CancelSubscription godoc
// @Summary      Cancel subscription
// @Description  Cancels the user's active subscription if no slots are being used.
// @Tags         payment-service-subscriptions
// @Security     BearerAuth
// @Produce      json
// @Success      200      {object}  dtos.SubscriptionStatusResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /subscriptions [delete]
func (zc *subscriptionController) CancelSubscription(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		logger.Logger.Error("Failed to parse user_id from context: " + err.Error())
		_ = c.Error(custom_errors.NewBadRequestError("Invalid user_id in context"))
		return
	}

	sub, err := zc.zonesSubscriptionsService.CancelSubscription(userID)
	if err != nil {
		logger.Logger.Error("Failed to cancel subscription for user " + userID.String() + ": " + err.Error())
		_ = c.Error(custom_errors.NewBadRequestError("Failed to cancel subscription: " + err.Error()))
		return
	}

	amountDue, _ := sub.AmountDue.Float64()

	res := &dtos.SubscriptionStatusResponse{
		Slots:           sub.TotalSlots,
		UsedSlots:       sub.UsedSlots,
		Status:          string(sub.Status),
		NextBillingDate: sub.NextBillingDate,
		AmountDue:       amountDue,
	}
	common_handlers.HandleSuccessResponse(c, 200, res)
}

// GetStatus godoc
// @Summary      Get subscription status
// @Description  Retrieves the active subscription status for the user.
// @Tags         payment-service-subscriptions
// @Security     BearerAuth
// @Produce      json
// @Success      200      {object}  dtos.SubscriptionStatusResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Failure      404      {object}  dtos.ErrorResponse
// @Failure      500      {object}  dtos.ErrorResponse
// @Router       /subscriptions/status [get]
func (zc *subscriptionController) GetStatus(c *gin.Context) {
	userID, err := common_handlers.GetUserIDFromSession(c)
	if err != nil {
		_ = c.Error(custom_errors.NewUnauthorizedError(err.Error()))
		return
	}

	sub, err := zc.zonesSubscriptionsService.GetByUserID(userID)
	if err != nil {
		logger.Logger.Warn("User " + userID.String() + " does not have an active subscription")
		_ = c.Error(custom_errors.NewNotFoundError("User does not have an active subscription"))
		return
	}

	amountDue, _ := sub.AmountDue.Float64()

	res := &dtos.SubscriptionStatusResponse{
		Slots:           sub.TotalSlots,
		UsedSlots:       sub.UsedSlots,
		Status:          string(sub.Status),
		NextBillingDate: sub.NextBillingDate,
		AmountDue:       amountDue,
	}
	common_handlers.HandleSuccessResponse(c, 200, res)
}

// CheckInternalAvailability godoc
// @Summary      Internal slots availability check
// @Description  Internal endpoint used by world-service to verify available spots.
// @Tags         payment-service-subscriptions
// @Security     ServerFixedToken
// @Produce      json
// @Param        user_id  path      string  true  "User ID"
// @Success      200      {object}  dtos.InternalSlotsCheckResponse
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Router       /subscriptions/internal/users/{user_id}/status [get]
func (zc *subscriptionController) CheckInternalAvailability(c *gin.Context) {
	userIdStr := c.Param("user_id")
	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Logger.Error("Failed to parse user_id from param: " + err.Error())
		_ = c.Error(custom_errors.NewBadRequestError("Invalid user_id in path"))
		return
	}

	allowed, freeSlots, err := zc.zonesSubscriptionsService.CheckAvalibility(userID)
	if err != nil {
		logger.Logger.Warn("User " + userID.String() + " does not have an active subscription: " + err.Error())
		_ = c.Error(custom_errors.NewUnauthorizedError("User does not have an active subscription"))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, dtos.InternalSlotsCheckResponse{
		Allowed:   allowed,
		FreeSlots: freeSlots,
	})
}

// InternalUpdateUsedSlots godoc
// @Summary      Internal modify used slots count
// @Description  Internal endpoint used by world-service to add or remove consumed slots for a user.
// @Tags         payment-service-subscriptions
// @Security     ServerFixedToken
// @Accept       json
// @Produce      json
// @Param        user_id  path      string                               true  "User ID"
// @Param        request  body      dtos.InternalUpdateUsedSlotsRequest  true  "Slots payload"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  dtos.ErrorResponse
// @Failure      401      {object}  dtos.ErrorResponse
// @Router       /subscriptions/internal/users/{user_id}/used-slots [put]
func (zc *subscriptionController) InternalUpdateUsedSlots(c *gin.Context) {
	userIdStr := c.Param("user_id")
	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Logger.Error("Failed to parse user_id from param: " + err.Error())
		_ = c.Error(custom_errors.NewBadRequestError("Invalid user_id in path"))
		return
	}

	var req dtos.InternalUpdateUsedSlotsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(custom_errors.NewBadRequestError("invalid request body: " + err.Error()))
		return
	}

	if err := zc.zonesSubscriptionsService.UpdateUsedSlots(userID, req.Slots, req.AreUsed); err != nil {
		logger.Logger.Error("Failed to update used slots: " + err.Error())
		_ = c.Error(custom_errors.NewBadRequestError(err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, gin.H{"status": "ok"})
}

// HandleWebhook godoc
// @Summary      Handle Stripe webhook for Subscriptions
// @Description  Processes Stripe webhook events for subscriptions changes/cancellations.
// @Tags         payment-service-subscriptions
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.WebhookResponse
// @Failure      400  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /subscriptions/webhook/stripe [post]
func (zc *subscriptionController) HandleWebhook(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(65536))

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		_ = c.Error(custom_errors.NewBadRequestError("failed to read request body"))
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		_ = c.Error(custom_errors.NewBadRequestError("missing Stripe-Signature header"))
		return
	}

	if err := zc.zonesSubscriptionsService.HandleWebhook(body, signature); err != nil {
		_ = c.Error(custom_errors.NewInternalServerError("cannot process webhook: " + err.Error()))
		return
	}

	common_handlers.HandleSuccessResponse(c, 200, &dtos.WebhookResponse{})
}

// GetPricingInfo godoc
// @Summary      Get price metrics and dates
// @Description  Return the price per slot alongside the next system-wide billing date.
// @Tags         payment-service-subscriptions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.PricingInfoResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /subscriptions/pricing [get]
func (zc *subscriptionController) GetPricingInfo(c *gin.Context) {
	price, nextBilling := zc.zonesSubscriptionsService.GetPricingInfo()

	res := &dtos.PricingInfoResponse{
		PricePerSlot:    price,
		NextBillingDate: nextBilling,
	}

	common_handlers.HandleSuccessResponse(c, 200, res)
}
