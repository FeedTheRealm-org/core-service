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
	email_sender "github.com/FeedTheRealm-org/core-service/internal/utils/email_sender"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v85"
)

var testConf *config.Config
var testDB *config.DB
var testSvc *zoneSubscriptionService
var testRepo zones_subscriptions_repo.ZonesSubscriptionsRepository
var testEmailServer email_sender.EmailSenderService

func TestMain(m *testing.M) {
	testConf = config.CreateConfig()
	logger.InitLogger(false)
	var err error
	testDB, err = config.NewDB(testConf)
	if err != nil {
		panic(err)
	}
	testRepo = zones_subscriptions_repo.NewSubscriptionRepository(testConf, testDB)
	testEmailServer = email_sender.NewEmailSenderService(testConf)
	testSvc = NewSubscriptionService(testConf, testRepo, testEmailServer).(*zoneSubscriptionService)

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

	sub := &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_pending",
		TotalSlots:       4,
		AmountDue:        decimal.Zero,
		Status:           "pending",
	}

	createZoneSubscription(t, sub)

	amount, err := testSvc.getNextInvoiceAmount(sub)
	assert.NoError(t, err)
	assert.True(t, amount.Equal(decimal.Zero))

	updated, _ := testRepo.GetByUserID(userID)
	updated.Status = stripe.SubscriptionStatusCanceled
	_, _ = testRepo.Update(updated)

	amount, err = testSvc.getNextInvoiceAmount(sub)
	assert.NoError(t, err)
	assert.True(t, amount.Equal(decimal.Zero))
}

func TestStopAllJobs_ResetsSlots(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	setupStopJobsServer(t, userID)

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_stop",
		TotalSlots:       5,
		UsedSlots:        2,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
	})

	sub, _ := testRepo.GetByUserID(userID)
	err := testSvc.stopAllJobs(sub)
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

func setupStopJobsServer(t *testing.T, userID uuid.UUID) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/world/internal/users/" + userID.String() + "/stop-jobs"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server url: %v", err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("failed to parse server port: %v", err)
	}

	originalPort := testConf.Server.Port
	testConf.Server.Port = port

	t.Cleanup(func() {
		server.Close()
		testConf.Server.Port = originalPort
	})
}

func TestAdminCreateSubscription_InvalidSlots(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	sub, err := testSvc.AdminCreateSubscription(userID, "user@example.com", 0)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "slots must be greater than 0")
}

func TestAdminCreateSubscription_NewUser(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	sub, err := testSvc.AdminCreateSubscription(userID, "user@example.com", 7)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, userID, sub.UserID)
	assert.Equal(t, 7, sub.TotalSlots)
	assert.Equal(t, 0, sub.UsedSlots)
	assert.True(t, sub.IsAdminGranted)
	assert.Equal(t, stripe.SubscriptionStatusActive, sub.Status)
	assert.True(t, sub.AmountDue.Equal(decimal.Zero))
	assert.Empty(t, sub.StripeCustomerID)
}

func TestAdminCreateSubscription_AlreadyActive(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_active",
		TotalSlots:       3,
		UsedSlots:        1,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
	})

	sub, err := testSvc.AdminCreateSubscription(userID, "user@example.com", 5)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "already has an active subscription")
}

func TestAdminCreateSubscription_AlreadyPendingCancellation(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_pending_cancel",
		TotalSlots:       3,
		UsedSlots:        1,
		AmountDue:        decimal.Zero,
		Status:           StatusPendingCancellation,
	})

	sub, err := testSvc.AdminCreateSubscription(userID, "user@example.com", 5)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "already has an active subscription")
}

func TestAdminCreateSubscription_ReusesInactiveRecord(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:               userID,
		StripeCustomerID:     "cust_old",
		StripeSubscriptionID: "sub_old",
		TotalSlots:           3,
		UsedSlots:            1,
		AmountDue:            decimal.NewFromFloat(2.5),
		Status:               stripe.SubscriptionStatusCanceled,
	})

	sub, err := testSvc.AdminCreateSubscription(userID, "user@example.com", 10)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, 10, sub.TotalSlots)
	assert.Equal(t, 0, sub.UsedSlots)
	assert.True(t, sub.IsAdminGranted)
	assert.Empty(t, sub.StripeCustomerID)
	assert.Empty(t, sub.StripeSubscriptionID)
	assert.Equal(t, stripe.SubscriptionStatusActive, sub.Status)
	assert.True(t, sub.AmountDue.Equal(decimal.Zero))

	persisted, err := testRepo.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, 10, persisted.TotalSlots)
}

func TestAdminUpdateSlots_InvalidSlots(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	sub, err := testSvc.AdminUpdateSlots(userID, 0)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "slots must be greater than 0")
}

func TestAdminUpdateSlots_NotFound(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	sub, err := testSvc.AdminUpdateSlots(userID, 5)
	assert.Error(t, err)
	assert.Nil(t, sub)
}

func TestAdminUpdateSlots_NotActive(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_inactive",
		TotalSlots:       5,
		UsedSlots:        1,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusCanceled,
		IsAdminGranted:   true,
	})

	sub, err := testSvc.AdminUpdateSlots(userID, 8)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "subscription is not active")
}

func TestAdminUpdateSlots_BelowUsedSlots(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_below",
		TotalSlots:       5,
		UsedSlots:        4,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
		IsAdminGranted:   true,
	})

	sub, err := testSvc.AdminUpdateSlots(userID, 2)
	assert.Error(t, err)
	assert.Nil(t, sub)
	_, isCannotExceed := err.(*CannotExceedTotalSlotsError)
	assert.True(t, isCannotExceed)
}

func TestAdminUpdateSlots_NotAdminGranted(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_notadmin",
		TotalSlots:       5,
		UsedSlots:        1,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
		IsAdminGranted:   false,
	})

	sub, err := testSvc.AdminUpdateSlots(userID, 8)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "cannot update slots for non-admin granted subscription")
}

func TestAdminUpdateSlots_Success(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_success",
		TotalSlots:       5,
		UsedSlots:        2,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
		IsAdminGranted:   true,
	})

	sub, err := testSvc.AdminUpdateSlots(userID, 12)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, 12, sub.TotalSlots)

	persisted, _ := testRepo.GetByUserID(userID)
	assert.Equal(t, 12, persisted.TotalSlots)
}

func TestAdminCancelSubscription_NotFound(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	sub, err := testSvc.AdminCancelSubscription(userID)
	assert.Error(t, err)
	assert.Nil(t, sub)
}

func TestAdminCancelSubscription_NotActive(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_notactive_cancel",
		TotalSlots:       5,
		UsedSlots:        1,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusCanceled,
		IsAdminGranted:   true,
	})

	sub, err := testSvc.AdminCancelSubscription(userID)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "subscription is not active")
}

func TestAdminCancelSubscription_AdminGranted_Immediate(t *testing.T) {
	clearZonesTables()
	userID := uuid.New()

	setupStopJobsServer(t, userID)

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           userID,
		StripeCustomerID: "cust_admin_cancel",
		TotalSlots:       5,
		UsedSlots:        2,
		AmountDue:        decimal.NewFromFloat(3.0),
		Status:           stripe.SubscriptionStatusActive,
		IsAdminGranted:   true,
	})

	sub, err := testSvc.AdminCancelSubscription(userID)
	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, stripe.SubscriptionStatusCanceled, sub.Status)
	assert.False(t, sub.IsAdminGranted)
	assert.Equal(t, 0, sub.TotalSlots)
	assert.Equal(t, 0, sub.UsedSlots)
	assert.True(t, sub.AmountDue.Equal(decimal.Zero))
	assert.Empty(t, sub.StripeCustomerID)
	assert.Empty(t, sub.StripeSubscriptionID)
}

func TestGetAllSubscriptions_Passthrough(t *testing.T) {
	clearZonesTables()

	for i := 0; i < 3; i++ {
		createZoneSubscription(t, &models.ZonesSubscriptions{
			UserID:           uuid.New(),
			StripeCustomerID: "cust_list_" + strconv.Itoa(i),
			TotalSlots:       5,
			UsedSlots:        0,
			AmountDue:        decimal.Zero,
			Status:           stripe.SubscriptionStatusActive,
		})
	}

	subs, total, err := testSvc.GetAllSubscriptions(0, 2)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, total)
	assert.Len(t, subs, 2)
}

func TestGetAllSubscriptions_DefaultsLimit(t *testing.T) {
	clearZonesTables()

	for i := 0; i < 3; i++ {
		createZoneSubscription(t, &models.ZonesSubscriptions{
			UserID:           uuid.New(),
			StripeCustomerID: "cust_default_" + strconv.Itoa(i),
			TotalSlots:       5,
			UsedSlots:        0,
			AmountDue:        decimal.Zero,
			Status:           stripe.SubscriptionStatusActive,
		})
	}

	subs, total, err := testSvc.GetAllSubscriptions(0, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, total)
	assert.Len(t, subs, 3)
}

func TestGetAllSubscriptions_DefaultsOffset(t *testing.T) {
	clearZonesTables()

	createZoneSubscription(t, &models.ZonesSubscriptions{
		UserID:           uuid.New(),
		StripeCustomerID: "cust_offset",
		TotalSlots:       5,
		UsedSlots:        0,
		AmountDue:        decimal.Zero,
		Status:           stripe.SubscriptionStatusActive,
	})

	subs, total, err := testSvc.GetAllSubscriptions(-5, 10)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, total)
	assert.Len(t, subs, 1)
}

func TestGetAllSubscriptions_Empty(t *testing.T) {
	clearZonesTables()

	subs, total, err := testSvc.GetAllSubscriptions(0, 10)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, total)
	assert.Empty(t, subs)
}
