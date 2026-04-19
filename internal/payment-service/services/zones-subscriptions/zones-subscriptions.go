package zones_subscriptions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	zones_subscriptions "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/zones-subscriptions"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	stripe "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/customer"
	"github.com/stripe/stripe-go/v84/invoice"
	"github.com/stripe/stripe-go/v84/subscription"
	"github.com/stripe/stripe-go/v84/subscriptionitem"
	"github.com/stripe/stripe-go/v84/webhook"
)

type CannotExceedTotalSlotsError struct{}

func (e *CannotExceedTotalSlotsError) Error() string {
	return "used slots cannot exceed total slots"
}

type zoneSubscriptionService struct {
	conf *config.Config
	repo zones_subscriptions.ZonesSubscriptionsRepository
}

func NewSubscriptionService(conf *config.Config, repo zones_subscriptions.ZonesSubscriptionsRepository) SubscriptionService {
	stripe.Key = conf.Stripe.StripeApiKey
	return &zoneSubscriptionService{conf: conf, repo: repo}
}

func (zs *zoneSubscriptionService) CreateCheckoutSession(userID uuid.UUID, slots int, successURL string, cancelURL string) (string, error) {
	logger.Logger.Infof("Creating checkout session for user %s (%d slots)", userID, slots)

	sub, err := zs.repo.GetByUserID(userID)
	if err == nil && sub != nil && sub.Status == stripe.SubscriptionStatusActive {
		logger.Logger.Warnf("User %s already has an active subscription", userID)
		return "", err
	}

	sub, err = zs.ensureCustomer(userID, slots, sub)
	if err != nil {
		return "", err
	}

	stripeParamsPriceData := &stripe.CheckoutSessionLineItemPriceDataParams{
		Currency: stripe.String("usd"),
		ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name: stripe.String("Zone"),
		},
		UnitAmount: stripe.Int64(500),
		Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
	}

	stripeParamsSession := &stripe.CheckoutSessionLineItemParams{
		PriceData: stripeParamsPriceData,
		Quantity:  stripe.Int64((int64)(slots)),
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems:  []*stripe.CheckoutSessionLineItemParams{stripeParamsSession},
		Customer:   stripe.String(sub.StripeCustomerID),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			BillingCycleAnchor: stripe.Int64(zs.nextBillingDate().Unix()),
			ProrationBehavior:  stripe.String("create_prorations"),
			Metadata: map[string]string{
				"user_id": userID.String(),
			},
		},
		ClientReferenceID:   stripe.String(userID.String()),
		AllowPromotionCodes: stripe.Bool(false),
	}

	sess, err := session.New(params)
	if err != nil {
		logger.Logger.Errorf("Failed to create Stripe Checkout session for user %s: %v", userID, err)
		return "", err
	}

	logger.Logger.Infof("Checkout session created for user %s", userID)
	return sess.URL, nil
}

func (zs *zoneSubscriptionService) UpdateSlots(userID uuid.UUID, newSlots int) (*models.ZonesSubscriptions, error) {
	logger.Logger.Infof("Updating subscription slots for user %s to %d", userID, newSlots)

	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		logger.Logger.Errorf("Subscription not found for user %s: %v", userID, err)
		return nil, err
	}

	if sub.StripeSubscriptionID == "" {
		return nil, fmt.Errorf("subscription not found for user")
	}
	if sub.Status != stripe.SubscriptionStatusActive {
		return nil, fmt.Errorf("subscription is not active")
	}

	if newSlots < sub.UsedSlots {
		return nil, &CannotExceedTotalSlotsError{}
	}

	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		logger.Logger.Errorf("Failed to fetch Stripe subscription %s: %v", sub.StripeSubscriptionID, err)
		return nil, err
	}

	if len(stripeSub.Items.Data) == 0 {
		return nil, err
	}

	_, err = subscriptionitem.Update(stripeSub.Items.Data[0].ID, &stripe.SubscriptionItemParams{
		Quantity:          stripe.Int64(int64(newSlots)),
		ProrationBehavior: stripe.String("create_prorations"),
	})
	if err != nil {
		logger.Logger.Errorf("Failed to update Stripe subscription item for user %s: %v", userID, err)
		return nil, err
	}

	logger.Logger.Infof("Successfully updated slots to %d for user %s", newSlots, userID)
	return sub, nil
}

func (zs *zoneSubscriptionService) UpdateUsedSlots(userID uuid.UUID, slots int, areUsed bool) error {
	logger.Logger.Infof("Updating used slots for user %s, slots: %d, areUsed: %v", userID, slots, areUsed)

	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		logger.Logger.Errorf("Subscription not found for user %s: %v", userID, err)
		return err
	}

	if areUsed {
		sub.UsedSlots += slots
	} else {
		sub.UsedSlots -= slots
	}

	if sub.UsedSlots < 0 {
		sub.UsedSlots = 0
	}
	if sub.UsedSlots > sub.TotalSlots {
		return &CannotExceedTotalSlotsError{}
	}

	if _, err := zs.repo.Update(sub); err != nil {
		logger.Logger.Errorf("Failed to update used slots for user %s: %v", userID, err)
		return err
	}
	return nil
}

func (zs *zoneSubscriptionService) GetByUserID(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	amountDue, err := zs.getNextInvoiceAmount(userID)
	if err != nil {
		logger.Logger.Errorf("Failed to fetch upcoming invoice for user %s during subscription deletion: %v", userID, err)
		return nil, err
	}
	sub.AmountDue = amountDue

	if _, err = zs.repo.Update(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (zs *zoneSubscriptionService) CheckAvalibility(userID uuid.UUID) (bool, int, error) {
	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return false, 0, err
	}

	if sub.Status != stripe.SubscriptionStatusActive {
		return false, 0, nil
	}

	return (sub.TotalSlots - sub.UsedSlots) > 0, max(sub.TotalSlots-sub.UsedSlots, 0), nil
}

func (zs *zoneSubscriptionService) CancelSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	logger.Logger.Infof("Cancelling subscription for user %s", userID)

	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if sub.StripeSubscriptionID == "" {
		return nil, fmt.Errorf("subscription not found for user")
	}

	if sub.Status != stripe.SubscriptionStatusActive {
		return nil, fmt.Errorf("subscription is not active")
	}

	if sub.UsedSlots > 0 {
		return nil, fmt.Errorf("cannot cancel subscription because %d slots are currently in use. Please delete your zones first", sub.UsedSlots)
	}

	params := &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(true),
		Prorate:    stripe.Bool(true),
	}

	_, err = subscription.Cancel(sub.StripeSubscriptionID, params)
	if err != nil {
		logger.Logger.Errorf("Failed to cancel Stripe subscription %s for user %s: %v", sub.StripeSubscriptionID, userID, err)
		return nil, err
	}

	sub.Status = stripe.SubscriptionStatusCanceled
	if _, err = zs.repo.Update(sub); err != nil {
		logger.Logger.Errorf("Failed to update DB status to canceled for user %s: %v", userID, err)
		return nil, err
	}

	logger.Logger.Infof("Subscription for user %s cancelled", userID)
	return sub, nil
}

func (zs *zoneSubscriptionService) HandleWebhook(payload []byte, signature string) error {
	event, err := webhook.ConstructEvent(payload, signature, zs.conf.Stripe.StripeWebhookSecret)
	if err != nil {
		logger.Logger.Error("Failed to verify Stripe webhook signature: " + err.Error())
		return err
	}

	switch event.Type {
	case "customer.subscription.created":
		var stripeSub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
			return err
		}

		userIDStr, ok := stripeSub.Metadata["user_id"]
		if !ok || userIDStr == "" {
			return err
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return err
		}

		dbSub, err := zs.repo.GetByUserID(userID)
		if err != nil {
			return err
		}

		dbSub.StripeSubscriptionID = stripeSub.ID
		dbSub.Status = stripe.SubscriptionStatusActive
		dbSub.NextBillingDate = zs.nextBillingDate()
		dbSub.TotalSlots = int(stripeSub.Items.Data[0].Quantity)
		dbSub.AmountDue = decimal.NewFromInt(5)

		if _, err = zs.repo.Update(dbSub); err != nil {
			return err
		}

		logger.Logger.Infof("Handled subscription.created for user %s, stripe sub %s", userID, stripeSub.ID)
		return nil
	case "customer.subscription.updated":
		var stripeSub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
			return err
		}

		userIDStr, ok := stripeSub.Metadata["user_id"]
		if !ok || userIDStr == "" {
			return err
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return err
		}

		dbSub, err := zs.repo.GetByUserID(userID)
		if err != nil {
			return err
		}

		dbSub.StripeSubscriptionID = stripeSub.ID
		dbSub.Status = stripeSub.Status
		dbSub.NextBillingDate = zs.nextBillingDate()
		dbSub.TotalSlots = int(stripeSub.Items.Data[0].Quantity)

		amountDue, err := zs.getNextInvoiceAmount(dbSub.UserID)
		if err != nil {
			logger.Logger.Errorf("Failed to fetch upcoming invoice for user %s during subscription deletion: %v", dbSub.UserID, err)
			return err
		}
		dbSub.AmountDue = amountDue

		if _, err = zs.repo.Update(dbSub); err != nil {
			return err
		}

		logger.Logger.Infof("Handled subscription.updated for user %s", userID)
		return nil
	case "customer.subscription.deleted":
		var stripeSub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
			return err
		}

		dbSub, err := zs.repo.GetByStripeSubscriptionID(stripeSub.ID)
		if err != nil {
			return err
		}

		dbSub.Status = stripeSub.Status
		dbSub.TotalSlots = int(stripeSub.Items.Data[0].Quantity)

		amountDue, err := zs.getNextInvoiceAmount(dbSub.UserID)
		if err != nil {
			logger.Logger.Errorf("Failed to fetch upcoming invoice for user %s during subscription deletion: %v", dbSub.UserID, err)
			return err
		}
		dbSub.AmountDue = amountDue

		if _, err = zs.repo.Update(dbSub); err != nil {
			return err
		}

		logger.Logger.Infof("Handled subscription.deleted for stripe sub %s", stripeSub.ID)
		return nil
	default:
		logger.Logger.Warnf("Unhandled Stripe webhook event type: %s", event.Type)
		return nil
	}
}

func (zs *zoneSubscriptionService) getNextInvoiceAmount(userID uuid.UUID) (decimal.Decimal, error) {
	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return decimal.Zero, err
	}

	inv, err := invoice.CreatePreview(&stripe.InvoiceCreatePreviewParams{
		Subscription: stripe.String(sub.StripeSubscriptionID),
	})
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromInt(inv.AmountDue).Div(decimal.NewFromInt(100)), nil
}

func (zs *zoneSubscriptionService) ensureCustomer(
	userID uuid.UUID,
	slots int,
	existingSub *models.ZonesSubscriptions,
) (sub *models.ZonesSubscriptions, err error) {
	if existingSub != nil && existingSub.StripeCustomerID != "" {
		return existingSub, nil
	}

	c, err := customer.New(&stripe.CustomerParams{
		Metadata: map[string]string{
			"user_id": userID.String(),
		},
	})
	if err != nil {
		logger.Logger.Errorf("Failed to create Stripe customer for user %s: %v", userID, err)
		return nil, err
	}

	sub = &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: c.ID,
		TotalSlots:       slots,
		AmountDue:        decimal.NewFromFloat(zs.conf.Stripe.StripeZonePrice).Mul(decimal.NewFromInt(int64(slots))),
		Status:           "pending",
		NextBillingDate:  zs.nextBillingDate(),
	}
	if sub, err = zs.repo.Create(sub); err != nil {
		logger.Logger.Errorf("Failed to create subscription DB record for user %s: %v", userID, err)
		return nil, err
	}

	logger.Logger.Infof("Created Stripe customer %s for user %s", c.ID, userID)
	return sub, nil
}

func (zs *zoneSubscriptionService) nextBillingDate() time.Time {
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	anchorDay := 5

	now := time.Now().In(loc)

	candidate := time.Date(
		now.Year(),
		now.Month(),
		anchorDay,
		12, 0, 0, 0,
		loc,
	)

	if !candidate.After(now) {
		candidate = time.Date(
			now.Year(),
			now.Month()+1,
			anchorDay,
			12, 0, 0, 0,
			loc,
		)
	}

	return candidate.UTC()
}

func (zs *zoneSubscriptionService) GetPricingInfo() (float64, time.Time) {
	return zs.conf.Stripe.StripeZonePrice, zs.nextBillingDate()
}
