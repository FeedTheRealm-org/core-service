package creator_balances

import (
	creator_balances "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/creator-balances"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type creatorBalancesService struct {
	repo creator_balances.CreatorBalancesRepository
}

func NewCreatorBalancesService(repo creator_balances.CreatorBalancesRepository) CreatorBalancesService {
	return &creatorBalancesService{repo: repo}
}

func (s *creatorBalancesService) GetBalance(userId uuid.UUID) (decimal.Decimal, error) {
	return s.repo.GetBalance(userId)
}
