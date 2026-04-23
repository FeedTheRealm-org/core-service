package zones_subscriptions

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type ZonesSubscriptionsRepository interface {
	Create(subscription *models.ZonesSubscriptions) (*models.ZonesSubscriptions, error)
	Update(subscription *models.ZonesSubscriptions) (*models.ZonesSubscriptions, error)
	GetByUserID(userID uuid.UUID) (*models.ZonesSubscriptions, error)
	GetByStripeCustomerID(customerID string) (*models.ZonesSubscriptions, error)
	GetByStripeSubscriptionID(subID string) (*models.ZonesSubscriptions, error)
}
