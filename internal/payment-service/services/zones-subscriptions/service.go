package zones_subscriptions

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type SubscriptionService interface {
	UpdateSlots(userID uuid.UUID, newSlots int) (*models.ZonesSubscriptions, error)
	UpdateUsedSlots(userID uuid.UUID, slots int, areUsed bool) error
	GetAllSubscriptions(offset, limit int) ([]*models.ZonesSubscriptions, int64, error)
	GetByUserID(userID uuid.UUID) (*models.ZonesSubscriptions, error)
	GetPricingInfo() (float64, time.Time)
	CheckAvalibility(userID uuid.UUID) (bool, int, error)
	CreateCheckoutSession(userID uuid.UUID, email string, slots int, successUrl string, cancelUrl string) (string, error)
	CancelSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error)
	ReactivateSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error)
	AdminCreateSubscription(userID uuid.UUID, email string, slots int) (*models.ZonesSubscriptions, error)
	AdminUpdateSlots(userID uuid.UUID, newSlots int) (*models.ZonesSubscriptions, error)
	AdminCancelSubscription(userID uuid.UUID) (*models.ZonesSubscriptions, error)
	HandleWebhook(payload []byte, signature string) error
}
