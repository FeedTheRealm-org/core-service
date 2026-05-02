package creator_balances

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatorBalancesRepository interface {
	AddBalance(userId uuid.UUID, amount decimal.Decimal) error
	GetBalance(userId uuid.UUID) (decimal.Decimal, error)
}
