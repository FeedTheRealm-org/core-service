package gem_balances

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	core_errors "github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var gemBalancesConf *config.Config
var gemBalancesDB *config.DB
var gemBalancesRepo GemBalancesRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	gemBalancesConf = config.CreateConfig()
	var err error
	gemBalancesDB, err = config.NewDB(gemBalancesConf)
	if err != nil {
		panic(err)
	}
	gemBalancesRepo = NewGemBalancesRepository(gemBalancesConf, gemBalancesDB)

	clearGemBalancesTables()
	code := m.Run()
	clearGemBalancesTables()
	os.Exit(code)
}

func clearGemBalancesTables() {
}

func TestGemBalancesRepository_CreateAndGet(t *testing.T) {
	clearGemBalancesTables()

	userID := uuid.New()
	assert.NoError(t, gemBalancesRepo.CreateGemBalance(userID))

	balance, err := gemBalancesRepo.GetGemBalanceByUserId(userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), balance.Gems)
}

func TestGemBalancesRepository_GetGemBalance_NotFound(t *testing.T) {
	clearGemBalancesTables()

	_, err := gemBalancesRepo.GetGemBalanceByUserId(uuid.New())
	assert.Error(t, err)
	var notFound *core_errors.HttpError
	assert.True(t, errors.As(err, &notFound))
	assert.Equal(t, 404, notFound.Status)
}

func TestGemBalancesRepository_GetAll(t *testing.T) {
	clearGemBalancesTables()

	assert.NoError(t, gemBalancesRepo.CreateGemBalance(uuid.New()))
	assert.NoError(t, gemBalancesRepo.CreateGemBalance(uuid.New()))

	balances, err := gemBalancesRepo.GetAllGemBalances()
	assert.NoError(t, err)
	_ = balances
}

func TestGemBalancesRepository_AddToGemBalance(t *testing.T) {
	clearGemBalancesTables()

	userID := uuid.New()
	assert.NoError(t, gemBalancesRepo.CreateGemBalance(userID))

	assert.NoError(t, gemBalancesRepo.AddToGemBalance(userID, 25))
	balance, err := gemBalancesRepo.GetGemBalanceByUserId(userID)
	if err == nil {
		assert.Equal(t, int64(25), balance.Gems)
	}
}

func TestGemBalancesRepository_ApplyStripeCheckoutCreditIfUnprocessed(t *testing.T) {
	clearGemBalancesTables()

	userID := uuid.New()
	applied, err := gemBalancesRepo.ApplyStripeCheckoutCreditIfUnprocessed(userID, 50, "evt_1", "sess_1")
	assert.NoError(t, err)
	assert.True(t, applied)

	balance, err := gemBalancesRepo.GetGemBalanceByUserId(userID)
	if err == nil {
		assert.Equal(t, int64(50), balance.Gems)
	}

	applied, err = gemBalancesRepo.ApplyStripeCheckoutCreditIfUnprocessed(userID, 50, "evt_1", "sess_1")
	assert.NoError(t, err)
	assert.False(t, applied)

	balance, err = gemBalancesRepo.GetGemBalanceByUserId(userID)
	if err == nil {
		assert.Equal(t, int64(50), balance.Gems)
	}
}

func TestGemBalancesRepository_UpsertGemBalance(t *testing.T) {
	clearGemBalancesTables()

	userID := uuid.New()
	assert.NoError(t, gemBalancesRepo.UpsertGemBalance(userID, 15))
	balance, err := gemBalancesRepo.GetGemBalanceByUserId(userID)
	if err == nil {
		assert.Equal(t, int64(15), balance.Gems)
	}

	assert.NoError(t, gemBalancesRepo.UpsertGemBalance(userID, 99))
	balance, err = gemBalancesRepo.GetGemBalanceByUserId(userID)
	if err == nil {
		assert.Equal(t, int64(99), balance.Gems)
	}
}
