package zones_subscriptions

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type zonesSubscriptionsRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewSubscriptionRepository(conf *config.Config, db *config.DB) ZonesSubscriptionsRepository {
	return &zonesSubscriptionsRepository{conf: conf, db: db}
}

func (zsr *zonesSubscriptionsRepository) Create(subscription *models.ZonesSubscriptions) (*models.ZonesSubscriptions, error) {
	err := zsr.db.Conn.Create(subscription).Error
	if err != nil {
		return nil, err
	}
	return subscription, err
}

func (zsr *zonesSubscriptionsRepository) Update(subscription *models.ZonesSubscriptions) (*models.ZonesSubscriptions, error) {
	err := zsr.db.Conn.Save(subscription).Error
	if err != nil {
		return nil, err
	}
	return subscription, err
}

func (zsr *zonesSubscriptionsRepository) GetByUserID(userID uuid.UUID) (*models.ZonesSubscriptions, error) {
	var sub models.ZonesSubscriptions
	err := zsr.db.Conn.Where("user_id = ?", userID).First(&sub).Error
	return &sub, err
}

func (zsr *zonesSubscriptionsRepository) GetByStripeCustomerID(customerID string) (*models.ZonesSubscriptions, error) {
	var sub models.ZonesSubscriptions
	err := zsr.db.Conn.Where("stripe_customer_id = ?", customerID).First(&sub).Error
	return &sub, err
}

func (zsr *zonesSubscriptionsRepository) GetByStripeSubscriptionID(subID string) (*models.ZonesSubscriptions, error) {
	var sub models.ZonesSubscriptions
	err := zsr.db.Conn.Where("stripe_subscription_id = ?", subID).First(&sub).Error
	return &sub, err
}
