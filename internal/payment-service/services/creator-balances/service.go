package creator_balances

import "github.com/google/uuid"

type CreatorBalancesService interface {
	GetBalance(userId uuid.UUID) (float64, error)
}
