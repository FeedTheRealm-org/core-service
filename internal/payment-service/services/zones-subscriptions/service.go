package zones_subscriptions

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type SubscriptionService interface {
	UpdateSlots(userID uuid.UUID, newSlots int) (*models.ZonesSubscriptions, error)
	UpdateUsedSlots(userID uuid.UUID, slots int, areUsed bool) error
	GetByUserID(userID uuid.UUID) (*models.ZonesSubscriptions, error)
	CheckAvalibility(userID uuid.UUID) (bool, int, error)
	CreateCheckoutSession(userID uuid.UUID, slots int, successUrl string, cancelUrl string) (string, error)
	CancelSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error)
	HandleWebhook(payload []byte, signature string) error
}
