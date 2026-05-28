package zones_subscriptions

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func TestSubscriptionService_UpdateUsedSlots_ExceedsTotal(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()
	repo := &fakeZonesRepo{getByUserID: &models.ZonesSubscriptions{
		UserID:     userID,
		TotalSlots: 2,
		UsedSlots:  1,
		Status:     stripe.SubscriptionStatusActive,
	}}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	err := service.UpdateUsedSlots(userID, 2, true)
	assert.Error(t, err)
	_, exceeded := err.(*CannotExceedTotalSlotsError)
	assert.True(t, exceeded)
}

func TestSubscriptionService_UpdateUsedSlots_UpdateError(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()
	repo := &fakeZonesRepo{
		getByUserID: &models.ZonesSubscriptions{UserID: userID, TotalSlots: 5, UsedSlots: 1, Status: stripe.SubscriptionStatusActive},
		updateErr:   errors.New("boom"),
	}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	err := service.UpdateUsedSlots(userID, 1, true)
	assert.Error(t, err)
}

func TestSubscriptionService_CheckAvailability_Active(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()
	repo := &fakeZonesRepo{getByUserID: &models.ZonesSubscriptions{UserID: userID, TotalSlots: 5, UsedSlots: 2, Status: stripe.SubscriptionStatusActive}}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	allowed, freeSlots, err := service.CheckAvalibility(userID)
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 3, freeSlots)
}

func TestSubscriptionService_CheckAvailability_Inactive(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()
	repo := &fakeZonesRepo{getByUserID: &models.ZonesSubscriptions{UserID: userID, TotalSlots: 5, UsedSlots: 0, Status: stripe.SubscriptionStatusPastDue}}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	allowed, freeSlots, err := service.CheckAvalibility(userID)
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, freeSlots)
}

func TestSubscriptionService_GetByUserID_UpdateError(t *testing.T) {
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = true
	conf.Server.Environment = config.Development
	userID := uuid.New()
	repo := &fakeZonesRepo{
		getByUserID: &models.ZonesSubscriptions{
			UserID:               userID,
			StripeCustomerID:     "cust_123",
			StripeSubscriptionID: "",
			TotalSlots:           3,
			Status:               "pending",
			AmountDue:            decimal.Zero,
		},
		updateErr: errors.New("boom"),
	}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	_, err := service.GetByUserID(userID)
	assert.Error(t, err)
}

func TestSubscriptionService_StopAllJobs_HTTPError(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeZonesRepo{}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{}).(*zoneSubscriptionService)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	sub := &models.ZonesSubscriptions{UserID: uuid.New(), UsedSlots: 1}
	err := service.stopAllJobs(sub)
	assert.Error(t, err)
}

func TestSubscriptionService_StopAllJobs_ResetsSlots(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeZonesRepo{}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{}).(*zoneSubscriptionService)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	sub := &models.ZonesSubscriptions{UserID: uuid.New(), UsedSlots: 2}
	repo.updateErr = nil
	err := service.stopAllJobs(sub)
	assert.NoError(t, err)
	assert.True(t, repo.updateCalled)
	assert.Equal(t, 0, sub.UsedSlots)
}
