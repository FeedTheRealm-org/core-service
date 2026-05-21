package creator_balances

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatorBalancesService interface {
	GetBalance(userId uuid.UUID) (decimal.Decimal, error)
	GetAllBalances() ([]models.CreatorBalance, error)
}
