package zones_subscriptions

import (
	"fmt"
	"os"
	"testing"
	"time"

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

func TestZonesSubscriptionsRepository_GetAll(t *testing.T) {
	clearZonesTables()

	const total = 5
	created := make([]*models.ZonesSubscriptions, 0, total)

	for i := 0; i < total; i++ {
		sub := &models.ZonesSubscriptions{
			UserID:               uuid.New(),
			StripeCustomerID:     fmt.Sprintf("cust_%d", i),
			StripeSubscriptionID: fmt.Sprintf("sub_%d", i),
			TotalSlots:           5,
			UsedSlots:            0,
			AmountDue:            decimal.NewFromFloat(1.0),
			Status:               stripe.SubscriptionStatusActive,
		}
		c, err := zonesRepo.Create(sub)
		assert.NoError(t, err)
		created = append(created, c)
		time.Sleep(time.Millisecond)
	}

	t.Run("returns all when limit covers total", func(t *testing.T) {
		subs, count, err := zonesRepo.GetAll(0, total)
		assert.NoError(t, err)
		assert.EqualValues(t, total, count)
		assert.Len(t, subs, total)

		for i := 0; i < len(subs)-1; i++ {
			assert.True(t, subs[i].CreatedAt.Before(subs[i+1].CreatedAt) || subs[i].CreatedAt.Equal(subs[i+1].CreatedAt))
		}
		assert.Equal(t, created[0].StripeSubscriptionID, subs[0].StripeSubscriptionID)
	})

	t.Run("respects limit", func(t *testing.T) {
		subs, count, err := zonesRepo.GetAll(0, 2)
		assert.NoError(t, err)
		assert.EqualValues(t, total, count)
		assert.Len(t, subs, 2)
		assert.Equal(t, created[0].StripeSubscriptionID, subs[0].StripeSubscriptionID)
		assert.Equal(t, created[1].StripeSubscriptionID, subs[1].StripeSubscriptionID)
	})

	t.Run("respects offset", func(t *testing.T) {
		subs, count, err := zonesRepo.GetAll(2, 2)
		assert.NoError(t, err)
		assert.EqualValues(t, total, count)
		assert.Len(t, subs, 2)
		assert.Equal(t, created[2].StripeSubscriptionID, subs[0].StripeSubscriptionID)
		assert.Equal(t, created[3].StripeSubscriptionID, subs[1].StripeSubscriptionID)
	})

	t.Run("offset beyond total returns empty slice", func(t *testing.T) {
		subs, count, err := zonesRepo.GetAll(total+10, 10)
		assert.NoError(t, err)
		assert.EqualValues(t, total, count)
		assert.Empty(t, subs)
	})

	t.Run("empty table returns zero total and empty slice", func(t *testing.T) {
		clearZonesTables()
		subs, count, err := zonesRepo.GetAll(0, 10)
		assert.NoError(t, err)
		assert.EqualValues(t, 0, count)
		assert.Empty(t, subs)
	})
}
