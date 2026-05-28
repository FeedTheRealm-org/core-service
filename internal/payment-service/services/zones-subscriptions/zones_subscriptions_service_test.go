package zones_subscriptions

import (
	"errors"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/email_sender"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v85"
)

type fakeZonesRepo struct {
	getByUserID      *models.ZonesSubscriptions
	getByUserErr     error
	createCalled     bool
	createErr        error
	updateCalled     bool
	updateErr        error
	getByStripeSubID *models.ZonesSubscriptions
}

func (f *fakeZonesRepo) Create(subscription *models.ZonesSubscriptions) (*models.ZonesSubscriptions, error) {
	f.createCalled = true
	if f.createErr != nil {
		return nil, f.createErr
	}
	return subscription, nil
}

func (f *fakeZonesRepo) Update(subscription *models.ZonesSubscriptions) (*models.ZonesSubscriptions, error) {
	f.updateCalled = true
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return subscription, nil
}

func (f *fakeZonesRepo) GetByUserID(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	if f.getByUserErr != nil {
		return nil, f.getByUserErr
	}
	return f.getByUserID, nil
}

func (f *fakeZonesRepo) GetByStripeCustomerID(customerID string) (*models.ZonesSubscriptions, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeZonesRepo) GetByStripeSubscriptionID(subID string) (*models.ZonesSubscriptions, error) {
	return f.getByStripeSubID, nil
}

type fakeZonesEmailSender struct{}

func (f *fakeZonesEmailSender) CreateBaseEmailData(toEmail string) email_sender.BaseEmailData {
	return email_sender.BaseEmailData{ToEmail: toEmail}
}

func (f *fakeZonesEmailSender) SendPasswordResetEmail(data email_sender.PasswordResetEmailData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendVerificationEmail(data email_sender.VerificationEmailData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendGemPurchaseEmail(data email_sender.GemPurchaseEmailData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendGemPurchaseFailedEmail(data email_sender.GemPurchaseFailedEmailData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendSubscriptionStartedEmail(data email_sender.SubscriptionStartedData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendSubscriptionUpdatedEmail(data email_sender.SubscriptionUpdatedData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendPaymentRejectedEmail(data email_sender.SubscriptionPaymentRejectedData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendPaymentSuccessfulEmail(data email_sender.SubscriptionPaymentSuccessfulData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendSubscriptionCancelledEmail(data email_sender.SubscriptionCancelledData) error {
	return nil
}

func TestSubscriptionService_CreateCheckoutSession_ActiveSubscription(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeZonesRepo{getByUserID: &models.ZonesSubscriptions{Status: stripe.SubscriptionStatusActive}}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	url, err := service.CreateCheckoutSession(uuid.New(), "user@example.com", 2, "ok", "cancel")
	assert.Error(t, err)
	assert.Equal(t, "", url)
}

func TestSubscriptionService_EnsureCustomer_ReturnsExisting(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeZonesRepo{}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{}).(*zoneSubscriptionService)

	existing := &models.ZonesSubscriptions{StripeCustomerID: "cust_123", TotalSlots: 2}
	result, err := service.ensureCustomer(uuid.New(), "user@example.com", 2, existing)
	assert.NoError(t, err)
	assert.Equal(t, existing, result)
	assert.False(t, repo.createCalled)
}

func TestSubscriptionService_GetByUserID_PendingStatus(t *testing.T) {
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = true
	conf.Server.Environment = config.Development

	userID := uuid.New()
	repo := &fakeZonesRepo{getByUserID: &models.ZonesSubscriptions{
		UserID:               userID,
		StripeCustomerID:     "cust_123",
		StripeSubscriptionID: "",
		TotalSlots:           3,
		Status:               "pending",
		AmountDue:            decimal.Zero,
	}}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	sub, err := service.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, sub.UserID)
	assert.True(t, sub.AmountDue.GreaterThan(decimal.Zero))
}

func TestSubscriptionService_GetPricingInfo(t *testing.T) {
	conf := config.CreateConfig()
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{})

	price, next := service.GetPricingInfo()
	assert.True(t, price > 0)
	assert.True(t, next.After(time.Now().Add(-time.Minute)))
}
