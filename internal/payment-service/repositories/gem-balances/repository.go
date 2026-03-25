package gem_balances

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type GemBalancesRepository interface {
	CreateGemBalance(userId uuid.UUID) error
	GetAllGemBalances() ([]*models.GemBalance, error)
	GetGemBalanceByUserId(userId uuid.UUID) (*models.GemBalance, error)
	AddToGemBalance(userId uuid.UUID, gems int) error
	ApplyStripeCheckoutCreditIfUnprocessed(userId uuid.UUID, gems int, eventID string, sessionID string) (bool, error)
	UpdateGemBalance(userId uuid.UUID, gems int) error
}
