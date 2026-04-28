package creator_balances

import (
	"github.com/google/uuid"
)

type CreatorBalancesRepository interface {
	AddBalance(userId uuid.UUID, amount float64) error
	GetBalance(userId uuid.UUID) (float64, error)
}
