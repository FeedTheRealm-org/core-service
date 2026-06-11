package zones_subscriptions

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v85"
)

var zonesConf *config.Config
var zonesDB *config.DB
var zonesRepo ZonesSubscriptionsRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	zonesConf = config.CreateConfig()
	var err error
	zonesDB, err = config.NewDB(zonesConf)
	if err != nil {
		panic(err)
	}
	zonesRepo = NewSubscriptionRepository(zonesConf, zonesDB)

	clearZonesTables()
	code := m.Run()
	clearZonesTables()
	os.Exit(code)
}

func clearZonesTables() {
	_ = zonesDB.Conn.Exec("TRUNCATE TABLE zones_subscriptions RESTART IDENTITY CASCADE;")
}

func TestZonesSubscriptionsRepository_CreateGetUpdate(t *testing.T) {
	clearZonesTables()

	userID := uuid.New()
	sub := &models.ZonesSubscriptions{
		UserID:               userID,
		StripeCustomerID:     "cust_123",
		StripeSubscriptionID: "sub_123",
		TotalSlots:           5,
		UsedSlots:            2,
		AmountDue:            decimal.NewFromFloat(3.5),
		Status:               stripe.SubscriptionStatusActive,
	}

	created, err := zonesRepo.Create(sub)
	assert.NoError(t, err)
	assert.Equal(t, userID, created.UserID)

	byUser, err := zonesRepo.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, "cust_123", byUser.StripeCustomerID)

	byCustomer, err := zonesRepo.GetByStripeCustomerID("cust_123")
	assert.NoError(t, err)
	assert.Equal(t, userID, byCustomer.UserID)

	bySub, err := zonesRepo.GetByStripeSubscriptionID("sub_123")
	assert.NoError(t, err)
	assert.Equal(t, userID, bySub.UserID)

	byUser.UsedSlots = 4
	updated, err := zonesRepo.Update(byUser)
	assert.NoError(t, err)
	assert.Equal(t, 4, updated.UsedSlots)
}
