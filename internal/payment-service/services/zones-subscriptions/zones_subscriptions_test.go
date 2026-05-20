package zones_subscriptions

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	zones_subscriptions_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/zones-subscriptions"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v85"
)

var testConf *config.Config
var testDB *config.DB
var testSvc *zoneSubscriptionService
var testRepo zones_subscriptions_repo.ZonesSubscriptionsRepository

func TestMain(m *testing.M) {
	testConf = config.CreateConfig()
	logger.InitLogger(false)
	var err error
	testDB, err = config.NewDB(testConf)
	if err != nil {
		panic(err)
	}
	testRepo = zones_subscriptions_repo.NewSubscriptionRepository(testConf, testDB)
	testSvc = NewSubscriptionService(testConf, testRepo).(*zoneSubscriptionService)

	clearZonesTables()
	code := m.Run()
	clearZonesTables()
	os.Exit(code)
}

func clearZonesTables() {
	_ = testDB.Conn.Exec("TRUNCATE TABLE zones_subscriptions RESTART IDENTITY CASCADE;")
}

func createZoneSubscription(t *testing.T, sub *models.ZonesSubscriptions) {
	_, err := testRepo.Create(sub)
	if err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}
}

func TestUpdateUsedSlots_IncreaseDecreaseClamp(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_123",
		TotalSlots:       5,
		UsedSlots:        1,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
	})

	err := testSvc.UpdateUsedSlots(userID, 2, true)
	assert.NoError(t, err)
	updated, _ := testRepo.GetByUserID(userID)
	assert.Equal(t, 3, updated.UsedSlots)

	err = testSvc.UpdateUsedSlots(userID, 10, false)
	assert.NoError(t, err)
	updated, _ = testRepo.GetByUserID(userID)
	assert.Equal(t, 0, updated.UsedSlots)
}

func TestUpdateUsedSlots_ExceedsTotal(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_456",
		TotalSlots:       3,
		UsedSlots:        2,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
	})

	err := testSvc.UpdateUsedSlots(userID, 2, true)
	assert.Error(t, err)
	_, isCannotExceed := err.(*CannotExceedTotalSlotsError)
	assert.True(t, isCannotExceed)
	updated, _ := testRepo.GetByUserID(userID)
	assert.Equal(t, 2, updated.UsedSlots)
}

func TestCheckAvailability_StatusInactive(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_789",
		TotalSlots:       5,
		UsedSlots:        0,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusPastDue,
	})

	allowed, freeSlots, err := testSvc.CheckAvalibility(userID)
	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, freeSlots)
}

func TestCheckAvailability_Active(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_abc",
		TotalSlots:       5,
		UsedSlots:        2,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
	})

	allowed, freeSlots, err := testSvc.CheckAvalibility(userID)
	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 3, freeSlots)
}

func TestGetByUserID_SubscriptionOff(t *testing.T) {
	originalSubscriptionOn := testConf.Server.SubscriptionOn
	originalEnv := testConf.Server.Environment
	defer func() {
		testConf.Server.SubscriptionOn = originalSubscriptionOn
		testConf.Server.Environment = originalEnv
	}()

	testConf.Server.SubscriptionOn = false
	testConf.Server.Environment = config.Development

	userID := uuid.New()

	sub, err := testSvc.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, sub.UserID)
	assert.Equal(t, 1000, sub.TotalSlots)
	assert.Equal(t, 0, sub.UsedSlots)
	assert.Equal(t, stripe.SubscriptionStatusActive, sub.Status)
}

func TestGetNextInvoiceAmount_PendingOrCanceled(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_pending",
		TotalSlots:       4,
		AmountDue:        decimal.Zero,
		Status:           "pending",
	})

	amount, err := testSvc.getNextInvoiceAmount(userID)
	assert.NoError(t, err)
	assert.True(t, amount.Equal(decimal.NewFromFloat(testConf.Stripe.StripeZonePrice*4)))

	updated, _ := testRepo.GetByUserID(userID)
	updated.Status = stripe.SubscriptionStatusCanceled
	_, _ = testRepo.Update(updated)

	amount, err = testSvc.getNextInvoiceAmount(userID)
	assert.NoError(t, err)
	assert.True(t, amount.Equal(decimal.NewFromFloat(testConf.Stripe.StripeZonePrice*4)))
}

func TestStopAllJobs_ResetsSlots(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/world/internal/users/" + userID.String() + "/stop-jobs"
		if r.URL.Path != expectedPath {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server url: %v", err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("failed to parse server port: %v", err)
	}
	testConf.Server.Port = port

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_stop",
		TotalSlots:       5,
		UsedSlots:        2,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
	})

	sub, _ := testRepo.GetByUserID(userID)
	err = testSvc.stopAllJobs(sub)
	assert.NoError(t, err)
	updated, _ := testRepo.GetByUserID(userID)
	assert.Equal(t, 0, updated.UsedSlots)
}

func TestNextBillingDate_AnchorInFuture(t *testing.T) {
	anchorDay := time.Now().UTC().Day() + 1
	if anchorDay > 28 {
		anchorDay = 15
	}
	originalAnchor := testConf.Stripe.StripeBillingAnchorDay
	testConf.Stripe.StripeBillingAnchorDay = anchorDay
	defer func() {
		testConf.Stripe.StripeBillingAnchorDay = originalAnchor
	}()

	next := testSvc.nextBillingDate()
	assert.True(t, next.After(time.Now().UTC()))
}
