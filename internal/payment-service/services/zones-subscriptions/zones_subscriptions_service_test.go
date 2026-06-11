package zones_subscriptions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

type fakeZonesEmailSender struct {
	sendStartedCalled  bool
	sendPaidCalled     bool
	sendRejectedCalled bool
}

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
	f.sendStartedCalled = true
	return nil
}

func (f *fakeZonesEmailSender) SendSubscriptionUpdatedEmail(data email_sender.SubscriptionUpdatedData) error {
	return nil
}

func (f *fakeZonesEmailSender) SendPaymentRejectedEmail(data email_sender.SubscriptionPaymentRejectedData) error {
	f.sendRejectedCalled = true
	return nil
}

func (f *fakeZonesEmailSender) SendPaymentSuccessfulEmail(data email_sender.SubscriptionPaymentSuccessfulData) error {
	f.sendPaidCalled = true
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

func TestSubscriptionService_UpdateSlots_Validations(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()

	tests := []struct {
		name        string
		sub         *models.ZonesSubscriptions
		repoErr     error
		newSlots    int
		expectedErr string
	}{
		{
			name:        "Error: Subscription not found in DB",
			repoErr:     errors.New("db error"),
			expectedErr: "db error",
		},
		{
			name: "Error: Missing StripeSubscriptionID",
			sub: &models.ZonesSubscriptions{
				Status: stripe.SubscriptionStatusActive,
			},
			expectedErr: "subscription not found for user",
		},
		{
			name: "Error: Subscription not active",
			sub: &models.ZonesSubscriptions{
				StripeSubscriptionID: "sub_123",
				Status:               stripe.SubscriptionStatusCanceled,
			},
			expectedErr: "subscription is not active",
		},
		{
			name: "Error: New slots less than used slots",
			sub: &models.ZonesSubscriptions{
				StripeSubscriptionID: "sub_123",
				Status:               stripe.SubscriptionStatusActive,
				UsedSlots:            5,
			},
			newSlots:    4,
			expectedErr: "used slots cannot exceed total slots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeZonesRepo{getByUserID: tt.sub, getByUserErr: tt.repoErr}
			service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})
			_, err := service.UpdateSlots(userID, tt.newSlots)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestSubscriptionService_CancelSubscription_Validations(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()

	tests := []struct {
		name        string
		sub         *models.ZonesSubscriptions
		repoErr     error
		expectedErr string
	}{
		{
			name:        "Error: Subscription not found in DB",
			repoErr:     errors.New("db error"),
			expectedErr: "db error",
		},
		{
			name: "Error: Missing StripeSubscriptionID",
			sub: &models.ZonesSubscriptions{
				Status: stripe.SubscriptionStatusActive,
			},
			expectedErr: "subscription not found for user",
		},
		{
			name: "Error: Subscription not active",
			sub: &models.ZonesSubscriptions{
				StripeSubscriptionID: "sub_123",
				Status:               stripe.SubscriptionStatusCanceled,
			},
			expectedErr: "subscription is not active",
		},
		{
			name: "Error: Cannot cancel with used slots",
			sub: &models.ZonesSubscriptions{
				StripeSubscriptionID: "sub_123",
				Status:               stripe.SubscriptionStatusActive,
				UsedSlots:            1,
			},
			expectedErr: "cannot cancel subscription because 1 slots are currently in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeZonesRepo{getByUserID: tt.sub, getByUserErr: tt.repoErr}
			service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})
			_, err := service.CancelSubscription(userID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestSubscriptionService_HandleWebhook_InvalidSignature(t *testing.T) {
	conf := config.CreateConfig()
	conf.Stripe.StripeSubscriptionsWebhookSecret = "whsec_test"
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{})

	payload := []byte(`{"id":"evt_test","type":"customer.subscription.created"}`)
	signature := "t=123,v1=invalid_signature"

	err := service.HandleWebhook(payload, signature)

	assert.Error(t, err)
}

func generateStripeSignature(secret string, payload []byte) string {
	timestamp := time.Now().Unix()

	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("t=%d,v1=%s", timestamp, sig)
}

func buildStripeEventPayload(eventType string, dataObject interface{}) []byte {
	event := map[string]interface{}{
		"id":          "evt_test_12345",
		"object":      "event",
		"api_version": stripe.APIVersion,
		"type":        eventType,
		"created":     time.Now().Unix(),
		"data": map[string]interface{}{
			"object": dataObject,
		},
	}
	payload, _ := json.Marshal(event)
	return payload
}

func webhookConf(secret string) *config.Config {
	conf := config.CreateConfig()
	conf.Stripe.StripeSubscriptionsWebhookSecret = secret
	return conf
}

func TestSubscriptionService_HandleWebhook_SignatureError(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{})

	payload := []byte(`{"id":"evt_test","type":"customer.subscription.updated"}`)
	sig := "t=123,v1=firma_invalida_totalmente"

	err := service.HandleWebhook(payload, sig)

	assert.Error(t, err)
}

func TestSubscriptionService_HandleWebhook_UnhandledEvent(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{})

	obj := map[string]interface{}{
		"id": "sub_test",
	}
	payload := buildStripeEventPayload("evento.random.no.soportado", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)

	assert.NoError(t, err)
}

func TestSubscriptionService_HandleWebhook_SubscriptionUpdated(t *testing.T) {
	stripeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"id":"in_test", "amount_due": 0}`))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer stripeServer.Close()

	originalBackend := stripe.GetBackend(stripe.APIBackend)
	defer stripe.SetBackend(stripe.APIBackend, originalBackend)

	mockBackend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL: stripe.String(stripeServer.URL),
	})
	stripe.SetBackend(stripe.APIBackend, mockBackend)
	stripe.Key = "sk_test_dummy"

	secret := "whsec_test_secret"
	conf := webhookConf(secret)

	userID := uuid.New()
	sub := &models.ZonesSubscriptions{
		UserID:               userID,
		StripeSubscriptionID: "sub_123",
		TotalSlots:           5,
	}

	repo := &fakeZonesRepo{
		getByUserID:      sub,
		getByStripeSubID: sub,
	}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	obj := map[string]interface{}{
		"id":     "sub_123",
		"object": "subscription",
		"status": "active",
		"metadata": map[string]interface{}{
			"user_id": userID.String(),
		},
		"items": map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"quantity": 10,
				},
			},
		},
	}
	payload := buildStripeEventPayload("customer.subscription.updated", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)

	assert.NoError(t, err)
	assert.True(t, repo.updateCalled)
}

func TestSubscriptionService_HandleWebhook_SubscriptionDeleted(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)

	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	sub := &models.ZonesSubscriptions{
		UserID:               userID,
		StripeSubscriptionID: "sub_123",
		Status:               "active",
	}

	repo := &fakeZonesRepo{
		getByUserID:      sub,
		getByStripeSubID: sub,
	}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	obj := map[string]interface{}{
		"id":     "sub_123",
		"object": "subscription",
		"status": "canceled",
		"metadata": map[string]interface{}{
			"user_id": userID.String(),
		},
		"items": map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"quantity": 5,
				},
			},
		},
	}
	payload := buildStripeEventPayload("customer.subscription.deleted", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)

	assert.NoError(t, err)
	assert.True(t, repo.updateCalled)
}

func TestSubscriptionService_HandleWebhook_SubscriptionCreated(t *testing.T) {
	stripeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "/invoices/upcoming") {
			_, _ = w.Write([]byte(`{"id":"in_test", "amount_due": 1500}`))
		} else if strings.Contains(r.URL.Path, "/billing_portal/sessions") {
			_, _ = w.Write([]byte(`{"id":"bps_test", "url":"http://portal.fake"}`))
		} else {
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer stripeServer.Close()

	originalBackend := stripe.GetBackend(stripe.APIBackend)
	defer stripe.SetBackend(stripe.APIBackend, originalBackend)
	stripe.SetBackend(stripe.APIBackend, stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL: stripe.String(stripeServer.URL),
	}))
	stripe.Key = "sk_test_dummy"

	secret := "whsec_test_secret"
	conf := webhookConf(secret)

	userID := uuid.New()
	repo := &fakeZonesRepo{
		getByUserID: &models.ZonesSubscriptions{
			UserID:           userID,
			StripeCustomerID: "cus_123",
		},
	}
	emailSender := &fakeZonesEmailSender{}
	service := NewSubscriptionService(conf, repo, emailSender)

	obj := map[string]interface{}{
		"id":     "sub_new_123",
		"object": "subscription",
		"metadata": map[string]interface{}{
			"user_id": userID.String(),
			"email":   "test@example.com",
		},
		"items": map[string]interface{}{
			"data": []map[string]interface{}{
				{"quantity": 3},
			},
		},
	}
	payload := buildStripeEventPayload("customer.subscription.created", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)

	assert.NoError(t, err)
	assert.True(t, repo.updateCalled)
	assert.True(t, emailSender.sendStartedCalled)
}

func TestSubscriptionService_HandleWebhook_InvoicePaid(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)

	emailSender := &fakeZonesEmailSender{}
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, emailSender)

	obj := map[string]interface{}{
		"id":             "in_paid_123",
		"object":         "invoice",
		"customer_email": "success@example.com",
		"amount_paid":    2000,
		"created":        time.Now().Unix(),
		"lines": map[string]interface{}{
			"data": []map[string]interface{}{
				{"quantity": 2},
			},
		},
	}
	payload := buildStripeEventPayload("invoice.paid", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)

	assert.NoError(t, err)
	assert.True(t, emailSender.sendPaidCalled)
}

func TestSubscriptionService_HandleWebhook_InvoicePaymentFailed(t *testing.T) {
	stripeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"sub_fail_123", "status":"canceled"}`))
	}))
	defer stripeServer.Close()

	originalBackend := stripe.GetBackend(stripe.APIBackend)
	defer stripe.SetBackend(stripe.APIBackend, originalBackend)
	stripe.SetBackend(stripe.APIBackend, stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL: stripe.String(stripeServer.URL),
	}))
	stripe.Key = "sk_test_dummy"

	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer internalServer.Close()

	secret := "whsec_test_secret"
	conf := webhookConf(secret)
	portStr := strings.TrimPrefix(internalServer.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	repo := &fakeZonesRepo{
		getByStripeSubID: &models.ZonesSubscriptions{
			UserID:               uuid.New(),
			TotalSlots:           5,
			StripeSubscriptionID: "sub_fail_123",
		},
	}
	emailSender := &fakeZonesEmailSender{}
	service := NewSubscriptionService(conf, repo, emailSender)

	obj := map[string]interface{}{
		"id":             "in_fail_123",
		"object":         "invoice",
		"customer_email": "failed@example.com",
		"amount_due":     1500,
		"attempt_count":  1,
		"parent": map[string]interface{}{
			"subscription_details": map[string]interface{}{
				"subscription": map[string]interface{}{
					"id": "sub_fail_123",
				},
			},
		},
	}
	payload := buildStripeEventPayload("invoice.payment_failed", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)

	assert.NoError(t, err)
	assert.True(t, emailSender.sendRejectedCalled)
}

func TestSubscriptionService_GetByUserID_NotFound(t *testing.T) {
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = true
	conf.Server.Environment = config.Development

	repo := &fakeZonesRepo{getByUserErr: errors.New("not found")}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	_, err := service.GetByUserID(uuid.New())
	assert.Error(t, err)
}

func TestSubscriptionService_GetByUserID_CanceledUsesLocalCalc(t *testing.T) {
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = true
	conf.Server.Environment = config.Development

	userID := uuid.New()
	repo := &fakeZonesRepo{
		getByUserID: &models.ZonesSubscriptions{
			UserID:               userID,
			StripeCustomerID:     "cust_x",
			StripeSubscriptionID: "",
			TotalSlots:           2,
			Status:               stripe.SubscriptionStatusCanceled,
			AmountDue:            decimal.Zero,
		},
	}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	sub, err := service.GetByUserID(userID)
	assert.NoError(t, err)
	assert.True(t, sub.AmountDue.GreaterThanOrEqual(decimal.Zero))
}

// ─── CheckAvailability edge cases ────────────────────────────────────────────

func TestSubscriptionService_CheckAvailability_GetError(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeZonesRepo{getByUserErr: errors.New("db error")}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	allowed, slots, err := service.CheckAvalibility(uuid.New())
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, slots)
}

func TestSubscriptionService_CheckAvailability_NoFreeSlots(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()
	repo := &fakeZonesRepo{getByUserID: &models.ZonesSubscriptions{
		UserID:     userID,
		TotalSlots: 3,
		UsedSlots:  3,
		Status:     stripe.SubscriptionStatusActive,
	}}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	allowed, freeSlots, err := service.CheckAvalibility(userID)
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, freeSlots)
}

// ─── UpdateUsedSlots edge cases ───────────────────────────────────────────────

func TestSubscriptionService_UpdateUsedSlots_DecreaseClamp(t *testing.T) {
	conf := config.CreateConfig()
	userID := uuid.New()
	repo := &fakeZonesRepo{getByUserID: &models.ZonesSubscriptions{
		UserID:     userID,
		TotalSlots: 5,
		UsedSlots:  1,
		Status:     stripe.SubscriptionStatusActive,
	}}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	err := service.UpdateUsedSlots(userID, 10, false)
	assert.NoError(t, err)
	assert.Equal(t, 0, repo.getByUserID.UsedSlots)
}

func TestSubscriptionService_UpdateUsedSlots_GetError(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeZonesRepo{getByUserErr: errors.New("db error")}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	err := service.UpdateUsedSlots(uuid.New(), 1, true)
	assert.Error(t, err)
}

// ─── CancelSubscription ───────────────────────────────────────────────────────

func TestSubscriptionService_CancelSubscription_GetError(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeZonesRepo{getByUserErr: errors.New("db error")}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	_, err := service.CancelSubscription(uuid.New())
	assert.Error(t, err)
}

// ─── GetPricingInfo ───────────────────────────────────────────────────────────

func TestSubscriptionService_GetPricingInfo_NonZero(t *testing.T) {
	conf := config.CreateConfig()
	conf.Stripe.StripeZonePrice = 9.99
	conf.Stripe.StripeBillingAnchorDay = 15
	conf.Stripe.StripeBillingTimezone = "UTC"
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{})

	price, next := service.GetPricingInfo()
	assert.Equal(t, 9.99, price)
	assert.True(t, next.After(time.Now().Add(-time.Minute)))
}

// ─── nextBillingDate ─────────────────────────────────────────────────────────

func TestSubscriptionService_NextBillingDate_PastAnchorRollsOver(t *testing.T) {
	conf := config.CreateConfig()
	conf.Stripe.StripeBillingTimezone = "UTC"
	conf.Stripe.StripeBillingAnchorDay = 1
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{}).(*zoneSubscriptionService)

	next := service.nextBillingDate()
	assert.True(t, next.After(time.Now().UTC()))
	assert.Equal(t, 1, next.Day())
}

func TestSubscriptionService_NextBillingDate_InvalidTimezone(t *testing.T) {
	conf := config.CreateConfig()
	conf.Stripe.StripeBillingTimezone = "Invalid/Zone"
	conf.Stripe.StripeBillingAnchorDay = 15
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{}).(*zoneSubscriptionService)

	next := service.nextBillingDate()
	assert.True(t, next.After(time.Now().Add(-time.Minute)))
}

// ─── GetByUserID con SubscriptionOn=false en Production ──────────────────────

func TestSubscriptionService_GetByUserID_SubscriptionOffProduction(t *testing.T) {
	conf := config.CreateConfig()
	conf.Server.SubscriptionOn = false
	conf.Server.Environment = config.Production

	userID := uuid.New()
	repo := &fakeZonesRepo{
		getByUserID: &models.ZonesSubscriptions{
			UserID:               userID,
			StripeSubscriptionID: "",
			TotalSlots:           1,
			Status:               stripe.SubscriptionStatusCanceled,
			AmountDue:            decimal.Zero,
		},
	}
	service := NewSubscriptionService(conf, repo, &fakeZonesEmailSender{})

	sub, err := service.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, sub.UserID)
}

// ─── HandleWebhook — webhook con metadata faltante ───────────────────────────

func TestSubscriptionService_HandleWebhook_SubscriptionUpdated_MissingUserID(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{})

	obj := map[string]interface{}{
		"id":       "sub_123",
		"object":   "subscription",
		"status":   "active",
		"metadata": map[string]interface{}{},
		"items": map[string]interface{}{
			"data": []map[string]interface{}{{"quantity": 1}},
		},
	}
	payload := buildStripeEventPayload("customer.subscription.updated", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)
	assert.Error(t, err)
}

func TestSubscriptionService_HandleWebhook_SubscriptionCreated_MissingUserID(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, &fakeZonesEmailSender{})

	obj := map[string]interface{}{
		"id":       "sub_new",
		"object":   "subscription",
		"metadata": map[string]interface{}{},
		"items": map[string]interface{}{
			"data": []map[string]interface{}{{"quantity": 2}},
		},
	}
	payload := buildStripeEventPayload("customer.subscription.created", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)
	assert.Error(t, err)
}

func TestSubscriptionService_HandleWebhook_InvoicePaid_NoEmail(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)
	emailSender := &fakeZonesEmailSender{}
	service := NewSubscriptionService(conf, &fakeZonesRepo{}, emailSender)

	obj := map[string]interface{}{
		"id":             "in_noemail",
		"object":         "invoice",
		"customer_email": "",
		"amount_paid":    1000,
		"created":        time.Now().Unix(),
		"lines": map[string]interface{}{
			"data": []map[string]interface{}{{"quantity": 1}},
		},
	}
	payload := buildStripeEventPayload("invoice.paid", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)
	assert.NoError(t, err)
	assert.False(t, emailSender.sendPaidCalled)
}

func TestSubscriptionService_HandleWebhook_SubscriptionDeleted_NoEmail(t *testing.T) {
	secret := "whsec_test_secret"
	conf := webhookConf(secret)

	userID := uuid.New()
	sub := &models.ZonesSubscriptions{
		UserID:               userID,
		StripeSubscriptionID: "sub_del_noemail",
		Status:               "active",
	}
	repo := &fakeZonesRepo{
		getByUserID:      sub,
		getByStripeSubID: sub,
	}
	emailSender := &fakeZonesEmailSender{}
	service := NewSubscriptionService(conf, repo, emailSender)

	obj := map[string]interface{}{
		"id":       "sub_del_noemail",
		"object":   "subscription",
		"status":   "canceled",
		"metadata": map[string]interface{}{},
		"items": map[string]interface{}{
			"data": []map[string]interface{}{{"quantity": 1}},
		},
	}
	payload := buildStripeEventPayload("customer.subscription.deleted", obj)
	sig := generateStripeSignature(secret, payload)

	err := service.HandleWebhook(payload, sig)
	assert.NoError(t, err)
}
