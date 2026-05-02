package creator_balances

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatorBalancesService interface {
	GetBalance(userId uuid.UUID) (decimal.Decimal, error)
}
