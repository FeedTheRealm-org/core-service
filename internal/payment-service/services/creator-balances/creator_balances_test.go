package creator_balances

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	creator_balances_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/creator-balances"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var creatorBalancesConf *config.Config
var creatorBalancesDB *config.DB
var creatorBalancesRepo creator_balances_repo.CreatorBalancesRepository
var creatorBalancesSvc CreatorBalancesService

func TestMain(m *testing.M) {
	creatorBalancesConf = config.CreateConfig()
	logger.InitLogger(false)
	var err error
	creatorBalancesDB, err = config.NewDB(creatorBalancesConf)
	if err != nil {
		panic(err)
	}
	creatorBalancesRepo = creator_balances_repo.NewCreatorBalancesRepository(creatorBalancesConf, creatorBalancesDB)
	creatorBalancesSvc = NewCreatorBalancesService(creatorBalancesRepo)

	clearCreatorBalancesTables()
	code := m.Run()
	clearCreatorBalancesTables()
	os.Exit(code)
}

func clearCreatorBalancesTables() {
}

func TestCreatorBalances_GetBalance(t *testing.T) {
	clearCreatorBalancesTables()

	userID := uuid.New()
	balance, err := creatorBalancesSvc.GetBalance(userID)
	assert.NoError(t, err)
	assert.True(t, balance.Equal(decimal.Zero))

	err = creatorBalancesRepo.AddBalance(userID, decimal.NewFromFloat(12.5))
	assert.NoError(t, err)

	balance, err = creatorBalancesSvc.GetBalance(userID)
	assert.NoError(t, err)
	assert.True(t, balance.Equal(decimal.NewFromFloat(12.5)))
}

func TestCreatorBalances_GetAllBalances(t *testing.T) {
	clearCreatorBalancesTables()

	userA := uuid.New()
	userB := uuid.New()
	assert.NoError(t, creatorBalancesRepo.AddBalance(userA, decimal.NewFromFloat(2.5)))
	assert.NoError(t, creatorBalancesRepo.AddBalance(userB, decimal.NewFromFloat(3.5)))

	balances, err := creatorBalancesSvc.GetAllBalances()
	assert.NoError(t, err)
	assert.True(t, len(balances) >= 2)
}
