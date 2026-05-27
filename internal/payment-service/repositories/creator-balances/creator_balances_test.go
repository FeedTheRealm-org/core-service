package creator_balances

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var creatorBalancesConf *config.Config
var creatorBalancesDB *config.DB
var creatorBalancesRepo CreatorBalancesRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	creatorBalancesConf = config.CreateConfig()
	var err error
	creatorBalancesDB, err = config.NewDB(creatorBalancesConf)
	if err != nil {
		panic(err)
	}
	creatorBalancesRepo = NewCreatorBalancesRepository(creatorBalancesConf, creatorBalancesDB)

	clearCreatorBalancesTables()
	code := m.Run()
	clearCreatorBalancesTables()
	os.Exit(code)
}

func clearCreatorBalancesTables() {
}

func TestCreatorBalancesRepository_GetBalanceAndAdd(t *testing.T) {
	clearCreatorBalancesTables()

	userID := uuid.New()
	balance, err := creatorBalancesRepo.GetBalance(userID)
	assert.NoError(t, err)
	assert.True(t, balance.Equal(decimal.Zero))

	assert.NoError(t, creatorBalancesRepo.AddBalance(userID, decimal.NewFromFloat(12.5)))
	balance, err = creatorBalancesRepo.GetBalance(userID)
	assert.NoError(t, err)
	assert.True(t, balance.Equal(decimal.NewFromFloat(12.5)))
}

func TestCreatorBalancesRepository_GetAllBalances(t *testing.T) {
	clearCreatorBalancesTables()

	userID1 := uuid.New()
	userID2 := uuid.New()
	assert.NoError(t, creatorBalancesRepo.AddBalance(userID1, decimal.NewFromFloat(1)))
	assert.NoError(t, creatorBalancesRepo.AddBalance(userID2, decimal.NewFromFloat(2)))

	balances, err := creatorBalancesRepo.GetAllBalances()
	assert.NoError(t, err)
	found1 := false
	found2 := false
	for _, balance := range balances {
		if balance.UserID == userID1 {
			found1 = true
		}
		if balance.UserID == userID2 {
			found2 = true
		}
	}
	assert.True(t, found1)
	assert.True(t, found2)
	if len(balances) > 0 {
		assert.IsType(t, models.CreatorBalance{}, balances[0])
	}
}
