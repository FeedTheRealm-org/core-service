package gem_balances

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type GemBalancesService interface {
	// GetAllBalances retrieves all user balances.
	GetAllGemBalances() ([]*models.GemBalance, error)

	// GetBalanceByUserId retrieves the balance for a specific user.
	GetGemBalanceByUserId(userId uuid.UUID) (*models.GemBalance, error)

	// CreateBalance creates a new balance record for a user.
	CreateGemBalance(userId uuid.UUID) error

	// UpdateBalance updates the balance for a specific user.
	UpdateGemBalance(userId uuid.UUID, gems int64) error

	// PurchaseCosmetic processes the purchase of a cosmetic item using gems.
	PurchaseCosmetic(userId uuid.UUID, cosmeticId uuid.UUID) error

	// CreateCheckoutSession creates a new checkout session for a user.
	CreateCheckoutSession(userId uuid.UUID, packageId uuid.UUID, successUrl string, cancelUrl string) (string, error)

	// HandleWebhook processes incoming webhook events from the payment provider.
	HandleWebhook(payload []byte, signature string) error
}
