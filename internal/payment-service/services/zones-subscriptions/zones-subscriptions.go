package zones_subscriptions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	zones_subscriptions "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/zones-subscriptions"
	"github.com/FeedTheRealm-org/core-service/internal/utils/email_sender"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	stripe "github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/checkout/session"
	"github.com/stripe/stripe-go/v85/customer"
	stripe_invoice "github.com/stripe/stripe-go/v85/invoice"
	"github.com/stripe/stripe-go/v85/subscription"
	"github.com/stripe/stripe-go/v85/subscriptionitem"
	"github.com/stripe/stripe-go/v85/webhook"
)

type CannotExceedTotalSlotsError struct{}

func (e *CannotExceedTotalSlotsError) Error() string {
	return "used slots cannot exceed total slots"
}

type zoneSubscriptionService struct {
	conf        *config.Config
	repo        zones_subscriptions.ZonesSubscriptionsRepository
	emailSender email_sender.EmailSenderService
}

const DATE_FORMAT = "2006-01-02 15:04:05 MST"
const MIN_PRORATED_AMOUNT_CENTS = 60
const StatusPendingCancellation = "pending_cancellation"

func NewSubscriptionService(conf *config.Config, repo zones_subscriptions.ZonesSubscriptionsRepository, emailSender email_sender.EmailSenderService) SubscriptionService {
	stripe.Key = conf.Stripe.StripeApiKey
	return &zoneSubscriptionService{conf: conf, repo: repo, emailSender: emailSender}
}

func (zs *zoneSubscriptionService) calculateProratedAmount(slots int) int64 {
	pricePerSlot := int64(zs.conf.Stripe.StripeZonePrice * 100)

	now := time.Now().UTC()
	cycleStart := zs.currentBillingDate()

	nowUnix := now.Unix()
	billingEnd := zs.nextBillingDate().Unix()
	cycleStartUnix := cycleStart.Unix()

	ratio := float64(billingEnd-nowUnix) / float64(billingEnd-cycleStartUnix)
	totalAmount := pricePerSlot * int64(slots)

	return int64(math.Floor(float64(totalAmount) * ratio))
}

func (zs *zoneSubscriptionService) CreateCheckoutSession(userID uuid.UUID, email string, slots int, successURL string, cancelURL string) (string, error) {
	logger.Logger.Infof("Creating checkout session for user %s (%d slots)", userID, slots)

	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		sub = nil
	} else if sub.Status == stripe.SubscriptionStatusActive {
		return "", fmt.Errorf("user %s already has an active subscription", userID)
	}

	sub, err = zs.ensureCustomer(userID, email, slots, sub)
	if err != nil {
		return "", err
	}

	proratedAmount := zs.calculateProratedAmount(slots)

	stripeParamsPriceData := &stripe.CheckoutSessionLineItemPriceDataParams{
		Currency: stripe.String("usd"),
		ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name: stripe.String("Zone"),
		},
		UnitAmount: stripe.Int64(int64(zs.conf.Stripe.StripeZonePrice * 100)),
		Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
	}

	stripeParamsSession := &stripe.CheckoutSessionLineItemParams{
		PriceData: stripeParamsPriceData,
		Quantity:  stripe.Int64((int64)(slots)),
	}

	lineItems := []*stripe.CheckoutSessionLineItemParams{stripeParamsSession}

	if proratedAmount < MIN_PRORATED_AMOUNT_CENTS {
		adjustment := MIN_PRORATED_AMOUNT_CENTS - proratedAmount

		proratedItem := &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("usd"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String("Minimum charge adjustment"),
				},
				UnitAmount: stripe.Int64(adjustment),
			},
			Quantity: stripe.Int64(1),
		}

		lineItems = append(lineItems, proratedItem)
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems:  lineItems,
		Customer:   stripe.String(sub.StripeCustomerID),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			BillingCycleAnchor: stripe.Int64(zs.nextBillingDate().Unix()),
			ProrationBehavior:  stripe.String("create_prorations"),
			Metadata: map[string]string{
				"user_id": userID.String(),
				"email":   email,
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

	if sub.Status == StatusPendingCancellation {
		return nil, fmt.Errorf("cannot update slots for subscription pending cancellation")
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

func (zs *zoneSubscriptionService) GetAllSubscriptions(offset, limit int) ([]*models.ZonesSubscriptions, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	subs, total, err := zs.repo.GetAll(offset, limit)
	if err != nil {
		logger.Logger.Errorf("Failed to list subscriptions (offset=%d, limit=%d): %v", offset, limit, err)
		return nil, 0, err
	}

	return subs, total, nil
}

func (zs *zoneSubscriptionService) GetByUserID(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	if !zs.conf.Server.SubscriptionOn && zs.conf.Server.Environment != config.Production {
		logger.Logger.Infof("Subscription system is turned off, returning default subscription for user %s", userID)
		return &models.ZonesSubscriptions{
			UserID:     userID,
			TotalSlots: 1000,
			UsedSlots:  0,
			Status:     stripe.SubscriptionStatusActive,
		}, nil
	}

	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	amountDue, err := zs.getNextInvoiceAmount(sub)
	if err != nil {
		logger.Logger.Errorf("Failed to fetch upcoming invoice for user %s: %v", userID, err)
		return nil, err
	}

	sub.AmountDue = amountDue

	if _, err = zs.repo.Update(sub); err != nil {
		logger.Logger.Errorf("Failed to persist amount due for user %s: %v", userID, err)
		return nil, err
	}

	return sub, nil
}

func (zs *zoneSubscriptionService) CheckAvalibility(userID uuid.UUID) (bool, int, error) {
	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return false, 0, err
	}

	if sub.Status != stripe.SubscriptionStatusActive && sub.Status != StatusPendingCancellation {
		return false, 0, nil
	}

	return (sub.TotalSlots - sub.UsedSlots) > 0, max(sub.TotalSlots-sub.UsedSlots, 0), nil
}

func (zs *zoneSubscriptionService) CancelSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	logger.Logger.Infof("Scheduling cancellation at period end for user %s", userID)

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

	_, err = subscription.Update(sub.StripeSubscriptionID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	})
	if err != nil {
		logger.Logger.Errorf("Failed to schedule cancellation for Stripe subscription %s for user %s: %v", sub.StripeSubscriptionID, userID, err)
		return nil, err
	}

	sub.Status = StatusPendingCancellation
	if _, err = zs.repo.Update(sub); err != nil {
		logger.Logger.Errorf("Failed to update DB status to %s for user %s: %v", StatusPendingCancellation, userID, err)
		return nil, err
	}

	logger.Logger.Infof("Subscription for user %s scheduled to cancel at period end", userID)
	return sub, nil
}

func (zs *zoneSubscriptionService) ReactivateSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if sub.StripeSubscriptionID == "" {
		return nil, fmt.Errorf("subscription not found for user")
	}
	if sub.Status != StatusPendingCancellation {
		return nil, fmt.Errorf("subscription is not pending cancellation")
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	})
	if err != nil {
		logger.Logger.Errorf("Failed to undo cancellation for Stripe subscription %s for user %s: %v", sub.StripeSubscriptionID, userID, err)
		return nil, err
	}

	sub.Status = stripe.SubscriptionStatusActive
	if _, err = zs.repo.Update(sub); err != nil {
		logger.Logger.Errorf("Failed to update DB status to active for user %s: %v", userID, err)
		return nil, err
	}

	logger.Logger.Infof("Subscription for user %s reactivated (cancellation undone)", userID)
	return sub, nil
}

func (zs *zoneSubscriptionService) stopAllJobs(sub *models.ZonesSubscriptions) error {
	logger.Logger.Infof("Stopping all jobs for user %s", sub.UserID)

	if sub.UsedSlots != 0 {
		logger.Logger.Warnf("User %s has %d used slots during stopAllJobs, resetting to 0", sub.UserID, sub.UsedSlots)

		url := fmt.Sprintf("http://127.0.0.1:%d/world/internal/users/%s/stop-jobs", zs.conf.Server.Port, sub.UserID)
		resp, err := http.Get(url)
		if err != nil {
			logger.Logger.Errorf("Failed to send stop-jobs internal request for user %s: %v", sub.UserID, err)
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			logger.Logger.Errorf("Received non-OK status %d from stop-jobs internal request for user %s", resp.StatusCode, sub.UserID)
			return fmt.Errorf("failed to stop jobs for user %s, status code: %d", sub.UserID, resp.StatusCode)
		}

		sub.UsedSlots = 0
		if _, err := zs.repo.Update(sub); err != nil {
			logger.Logger.Errorf("Failed to reset used slots for user %s during stopAllJobs: %v", sub.UserID, err)
			return err
		}
	}

	return nil
}

func (zs *zoneSubscriptionService) AdminCreateSubscription(userID uuid.UUID, email string, slots int) (*models.ZonesSubscriptions, error) {
	logger.Logger.Infof("Admin granting comp subscription to user %s (%d slots)", userID, slots)

	if slots <= 0 {
		return nil, fmt.Errorf("slots must be greater than 0")
	}

	existing, err := zs.repo.GetByUserID(userID)
	if err == nil && existing != nil {
		if existing.Status == stripe.SubscriptionStatusActive || existing.Status == StatusPendingCancellation {
			return nil, fmt.Errorf("user %s already has an active subscription", userID)
		}

		existing.StripeCustomerID = ""
		existing.StripeSubscriptionID = ""
		existing.TotalSlots = slots
		existing.UsedSlots = 0
		existing.AmountDue = decimal.Zero
		existing.Status = stripe.SubscriptionStatusActive
		existing.IsAdminGranted = true
		existing.NextBillingDate = time.Time{}

		updated, err := zs.repo.Update(existing)
		if err != nil {
			logger.Logger.Errorf("Failed to update comp subscription for user %s: %v", userID, err)
			return nil, err
		}

		logger.Logger.Infof("Admin comp subscription (re-used record) active for user %s", userID)
		return updated, nil
	}

	sub := &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "",
		TotalSlots:       slots,
		UsedSlots:        0,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
		IsAdminGranted:   true,
		NextBillingDate:  time.Time{},
	}

	created, err := zs.repo.Create(sub)
	if err != nil {
		logger.Logger.Errorf("Failed to create comp subscription for user %s: %v", userID, err)
		return nil, err
	}

	logger.Logger.Infof("Admin comp subscription created for user %s", userID)
	return created, nil
}

func (zs *zoneSubscriptionService) AdminUpdateSlots(userID uuid.UUID, newSlots int) (*models.ZonesSubscriptions, error) {
	logger.Logger.Infof("Admin updating subscription slots for user %s to %d", userID, newSlots)

	if newSlots <= 0 {
		return nil, fmt.Errorf("slots must be greater than 0")
	}

	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		logger.Logger.Errorf("Subscription not found for user %s: %v", userID, err)
		return nil, err
	}

	if sub.Status != stripe.SubscriptionStatusActive && sub.Status != StatusPendingCancellation {
		return nil, fmt.Errorf("subscription is not active")
	}

	if newSlots < sub.UsedSlots {
		return nil, &CannotExceedTotalSlotsError{}
	}

	if !sub.IsAdminGranted {
		return nil, fmt.Errorf("cannot update slots for non-admin granted subscription")
	}

	sub.TotalSlots = newSlots

	updated, err := zs.repo.Update(sub)
	if err != nil {
		logger.Logger.Errorf("Failed to update comp subscription slots for user %s: %v", userID, err)
		return nil, err
	}

	logger.Logger.Infof("Admin comp subscription slots updated to %d for user %s", newSlots, userID)
	return updated, nil
}

func (zs *zoneSubscriptionService) AdminCancelSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	logger.Logger.Infof("Admin cancelling subscription for user %s", userID)

	sub, err := zs.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if sub.Status != stripe.SubscriptionStatusActive && sub.Status != StatusPendingCancellation {
		return nil, fmt.Errorf("subscription is not active")
	}

	if err := zs.stopAllJobs(sub); err != nil {
		logger.Logger.Errorf("Failed to run stopAllJobs for user %s: %v", userID, err)
		return nil, err
	}

	if !sub.IsAdminGranted {
		updated, err := zs.CancelSubscription(userID)
		if err != nil {
			logger.Logger.Errorf("Failed to schedule cancellation for user %s: %v", userID, err)
			return nil, err
		}
		return updated, nil
	}

	sub.Status = stripe.SubscriptionStatusCanceled
	sub.IsAdminGranted = false
	sub.TotalSlots = 0
	sub.UsedSlots = 0
	sub.AmountDue = decimal.Zero
	sub.StripeCustomerID = ""
	sub.StripeSubscriptionID = ""

	updated, err := zs.repo.Update(sub)
	if err != nil {
		logger.Logger.Errorf("Failed to cancel comp subscription for user %s: %v", userID, err)
		return nil, err
	}

	logger.Logger.Infof("Admin comp subscription cancelled immediately for user %s", userID)
	return updated, nil
}

func (zs *zoneSubscriptionService) HandleWebhook(payload []byte, signature string) error {
	event, err := webhook.ConstructEvent(payload, signature, zs.conf.Stripe.StripeSubscriptionsWebhookSecret)
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
			return fmt.Errorf("missing user_id in subscription metadata: %s", stripeSub.ID)
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return err
		}

		email, ok := stripeSub.Metadata["email"]
		if !ok || email == "" {
			return fmt.Errorf("missing email in subscription metadata: %s", stripeSub.ID)
		}

		dbSub, err := zs.repo.GetByUserID(userID)
		if err != nil {
			return err
		}

		dbSub.StripeSubscriptionID = stripeSub.ID
		dbSub.Status = stripe.SubscriptionStatusActive
		dbSub.NextBillingDate = zs.nextBillingDate()
		dbSub.TotalSlots = int(stripeSub.Items.Data[0].Quantity)

		amountDue, err := zs.getNextInvoiceAmount(dbSub)
		if err != nil {
			logger.Logger.Errorf("Failed to fetch upcoming invoice for user %s during subscription creation: %v", dbSub.UserID, err)
			return err
		}
		dbSub.AmountDue = amountDue

		if _, err = zs.repo.Update(dbSub); err != nil {
			return err
		}

		loc, err := time.LoadLocation(zs.conf.Stripe.StripeBillingTimezone)
		if err != nil {
			loc = time.UTC
		}

		portalURL, err := zs.createBillingPortalURL(dbSub.StripeCustomerID)
		if err != nil {
			logger.Logger.Errorf("Failed to create billing portal session: %v", err)
			portalURL = ""
		}

		float, _ := amountDue.Float64()

		err = zs.emailSender.SendSubscriptionStartedEmail(email_sender.SubscriptionStartedData{
			BaseEmailData:         zs.emailSender.CreateBaseEmailData(email),
			ZoneCount:             int64(dbSub.TotalSlots),
			Amount:                fmt.Sprintf("$%.2f", float),
			FirstBillingDate:      zs.nextBillingDate().In(loc).Format(DATE_FORMAT),
			ManageSubscriptionURL: portalURL,
		})
		if err != nil {
			logger.Logger.Error("Failed to send subscription started email for user " + dbSub.UserID.String() + ": " + err.Error())
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
			return fmt.Errorf("missing user_id in subscription metadata: %s", stripeSub.ID)
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

		if stripeSub.Status == stripe.SubscriptionStatusActive && stripeSub.CancelAtPeriodEnd {
			dbSub.Status = StatusPendingCancellation
		} else {
			dbSub.Status = stripeSub.Status
		}

		dbSub.NextBillingDate = zs.nextBillingDate()

		oldAmountDue := dbSub.AmountDue
		oldSlotsQuantity := dbSub.TotalSlots

		dbSub.AmountDue = decimal.NewFromFloat(zs.conf.Stripe.StripeZonePrice).Mul(decimal.NewFromInt(int64(stripeSub.Items.Data[0].Quantity)))
		dbSub.TotalSlots = int(stripeSub.Items.Data[0].Quantity)

		if stripeSub.Status == stripe.SubscriptionStatusActive {
			amountDue, err := zs.getNextInvoiceAmount(dbSub)
			if err != nil {
				logger.Logger.Errorf("Failed to fetch upcoming invoice for user %s during subscription update: %v", dbSub.UserID, err)
				return err
			}
			dbSub.AmountDue = amountDue

			email, emailOk := stripeSub.Metadata["email"]

			if emailOk && email != "" && dbSub.TotalSlots > 0 {
				oldAmountDueFloat, _ := oldAmountDue.Float64()
				NewAmountDueFloat, _ := dbSub.AmountDue.Float64()

				loc, err := time.LoadLocation(zs.conf.Stripe.StripeBillingTimezone)
				if err != nil {
					loc = time.UTC
				}

				err = zs.emailSender.SendSubscriptionUpdatedEmail(email_sender.SubscriptionUpdatedData{
					BaseEmailData:   zs.emailSender.CreateBaseEmailData(email),
					OldZoneCount:    int64(oldSlotsQuantity),
					OldAmount:       fmt.Sprintf("$%.2f", oldAmountDueFloat),
					NewZoneCount:    int64(dbSub.TotalSlots),
					NewAmount:       fmt.Sprintf("$%.2f", NewAmountDueFloat),
					NextBillingDate: zs.nextBillingDate().In(loc).Format(DATE_FORMAT),
				})
				if err != nil {
					logger.Logger.Error("Failed to send subscription updated email for user " + dbSub.UserID.String() + ": " + err.Error())
				}
			}
		}

		if _, err = zs.repo.Update(dbSub); err != nil {
			return err
		}

		logger.Logger.Infof("Handled subscription.updated for user %s, status: %s", userID, dbSub.Status)
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

		if err := zs.stopAllJobs(dbSub); err != nil {
			logger.Logger.Errorf("Failed to run stopAllJobs: %v", err)
			return err
		}

		email, emailOk := stripeSub.Metadata["email"]

		dbSub.Status = stripeSub.Status
		dbSub.StripeCustomerID = ""
		dbSub.StripeSubscriptionID = ""
		dbSub.TotalSlots = int(stripeSub.Items.Data[0].Quantity)
		dbSub.AmountDue = decimal.Zero

		if _, err = zs.repo.Update(dbSub); err != nil {
			return err
		}

		if emailOk && email != "" {
			err = zs.emailSender.SendSubscriptionCancelledEmail(email_sender.SubscriptionCancelledData{
				BaseEmailData: zs.emailSender.CreateBaseEmailData(email),
				ZoneCount:     dbSub.TotalSlots,
			})
			if err != nil {
				logger.Logger.Error("Failed to send subscription cancelled email for user " + dbSub.UserID.String() + ": " + err.Error())
				return err
			}
		} else {
			logger.Logger.Warnf("Missing email in subscription metadata for sub %s, skipping cancellation email", stripeSub.ID)
		}

		logger.Logger.Infof("Handled subscription.deleted for stripe sub %s, user %s", stripeSub.ID, dbSub.UserID)
		return nil
	case "invoice.paid":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return err
		}

		loc, err := time.LoadLocation(zs.conf.Stripe.StripeBillingTimezone)
		if err != nil {
			loc = time.UTC
		}

		if invoice.CustomerEmail != "" {
			err = zs.emailSender.SendPaymentSuccessfulEmail(email_sender.SubscriptionPaymentSuccessfulData{
				BaseEmailData:   zs.emailSender.CreateBaseEmailData(invoice.CustomerEmail),
				ZoneCount:       int(invoice.Lines.Data[0].Quantity),
				Amount:          fmt.Sprintf("$%.2f", float64(invoice.AmountPaid)/100),
				PaymentDate:     time.Unix(invoice.Created, 0).In(loc).Format(DATE_FORMAT),
				InvoiceID:       invoice.ID,
				NextBillingDate: zs.nextBillingDate().In(loc).Format(DATE_FORMAT),
			})
			if err != nil {
				logger.Logger.Error("Failed to send subscription payment successful email for invoice " + invoice.ID + ": " + err.Error())
				return err
			}
		} else {
			logger.Logger.Warnf("Missing CustomerEmail in invoice %s, skipping payment successful email", invoice.ID)
		}

		logger.Logger.Infof("Handled invoice.paid for invoice %s", invoice.ID)
		return nil
	case "invoice.payment_failed":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return err
		}

		var subscriptionId = invoice.Parent.SubscriptionDetails.Subscription.ID
		if subscriptionId == "" {
			logger.Logger.Warn("invoice.payment_failed received with no associated subscription, skipping")
			return nil
		}

		dbSub, err := zs.repo.GetByStripeSubscriptionID(subscriptionId)
		if err != nil {
			logger.Logger.Warn("invoice.payment_failed received for subscription ID with no DB record: " + subscriptionId)
			return err
		}

		declineCode := ""
		if invoice.Payments != nil {
			for _, p := range invoice.Payments.Data {
				if p.Payment != nil &&
					p.Payment.PaymentIntent != nil &&
					p.Payment.PaymentIntent.LastPaymentError != nil &&
					p.Status == "open" {
					declineCode = string(p.Payment.PaymentIntent.LastPaymentError.DeclineCode)
				}
			}
		}

		logger.Logger.Infof(
			"Payment failed for user %s, stripe sub %s, attempt #%d, decline code: %q",
			dbSub.UserID, subscriptionId, invoice.AttemptCount, declineCode,
		)

		if err := zs.stopAllJobs(dbSub); err != nil {
			logger.Logger.Errorf("Failed to run stopAllJobs: %v", err)
			return err
		}

		_, err = subscription.Cancel(subscriptionId, &stripe.SubscriptionCancelParams{
			InvoiceNow: stripe.Bool(false),
			Prorate:    stripe.Bool(false),
		})
		if err != nil {
			var stripeErr *stripe.Error
			if errors.As(err, &stripeErr) && stripeErr.HTTPStatusCode == 404 {
				logger.Logger.Infof("Stripe subscription %s not found, skipping cancel: %v", subscriptionId, err)
			} else {
				logger.Logger.Errorf("Failed to cancel Stripe subscription %s after payment failure: %v", subscriptionId, err)
				return err
			}
		}

		if invoice.CustomerEmail != "" {
			err = zs.emailSender.SendPaymentRejectedEmail(email_sender.SubscriptionPaymentRejectedData{
				BaseEmailData: zs.emailSender.CreateBaseEmailData(invoice.CustomerEmail),
				ZoneCount:     int64(dbSub.TotalSlots),
				Amount:        fmt.Sprintf("$%.2f", float64(invoice.AmountDue)/100),
			})
			if err != nil {
				logger.Logger.Error("Failed to send subscription payment rejected email for user " + dbSub.UserID.String() + ": " + err.Error())
				return err
			}
		} else {
			logger.Logger.Warnf("Missing CustomerEmail in invoice %s, skipping payment rejected email for user %s", invoice.ID, dbSub.UserID.String())
		}

		return nil
	default:
		logger.Logger.Infof("Unhandled Stripe webhook event type: %s", event.Type)
		return nil
	}
}

func (zs *zoneSubscriptionService) getNextInvoiceAmount(sub *models.ZonesSubscriptions) (decimal.Decimal, error) {
	if (sub.StripeSubscriptionID == "" && sub.Status == "pending") ||
		sub.Status == stripe.SubscriptionStatusCanceled ||
		sub.Status == StatusPendingCancellation {
		return decimal.Zero, nil
	}

	inv, err := stripe_invoice.CreatePreview(&stripe.InvoiceCreatePreviewParams{
		Subscription: stripe.String(sub.StripeSubscriptionID),
	})
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromInt(inv.AmountDue).Div(decimal.NewFromInt(100)), nil
}

func (zs *zoneSubscriptionService) ensureCustomer(
	userID uuid.UUID,
	email string,
	slots int,
	existingSub *models.ZonesSubscriptions,
) (sub *models.ZonesSubscriptions, err error) {
	if existingSub != nil && existingSub.StripeCustomerID != "" {
		return existingSub, nil
	}

	c, err := customer.New(&stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"user_id": userID.String(),
		},
	})
	if err != nil {
		logger.Logger.Errorf("Failed to create Stripe customer for user %s: %v", userID, err)
		return nil, err
	}

	if existingSub != nil {
		existingSub.StripeCustomerID = c.ID
		existingSub.TotalSlots = slots
		existingSub.AmountDue = decimal.NewFromFloat(zs.conf.Stripe.StripeZonePrice).Mul(decimal.NewFromInt(int64(slots)))
		existingSub.Status = "pending"
		existingSub.NextBillingDate = zs.nextBillingDate()
		if existingSub, err = zs.repo.Update(existingSub); err != nil {
			logger.Logger.Errorf("Failed to update subscription DB record for user %s: %v", userID, err)
			return nil, err
		}
		logger.Logger.Infof("Re-created Stripe customer %s for user %s", c.ID, userID)
		return existingSub, nil
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

func (zs *zoneSubscriptionService) currentBillingDate() time.Time {
	loc, err := time.LoadLocation(zs.conf.Stripe.StripeBillingTimezone)
	if err != nil {
		loc = time.UTC
	}
	anchorDay := zs.conf.Stripe.StripeBillingAnchorDay

	now := time.Now().In(loc)

	candidate := time.Date(
		now.Year(),
		now.Month(),
		anchorDay,
		12, 0, 0, 0,
		loc,
	)

	if candidate.After(now) {
		candidate = time.Date(
			now.Year(),
			now.Month()-1,
			anchorDay,
			12, 0, 0, 0,
			loc,
		)
	}

	return candidate.UTC()
}

func (zs *zoneSubscriptionService) nextBillingDate() time.Time {
	loc, err := time.LoadLocation(zs.conf.Stripe.StripeBillingTimezone)
	if err != nil {
		loc = time.UTC
	}
	anchorDay := zs.conf.Stripe.StripeBillingAnchorDay

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

func (zs *zoneSubscriptionService) createBillingPortalURL(stripeCustomerID string) (string, error) {
	sc := stripe.NewClient(zs.conf.Stripe.StripeApiKey, nil)
	params := &stripe.BillingPortalSessionCreateParams{
		Customer: stripe.String(stripeCustomerID),
	}
	result, err := sc.V1BillingPortalSessions.Create(context.TODO(), params)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (zs *zoneSubscriptionService) GetPricingInfo() (float64, time.Time) {
	return zs.conf.Stripe.StripeZonePrice, zs.nextBillingDate()
}
