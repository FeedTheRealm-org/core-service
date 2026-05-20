package gem_balances

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	creator_balances_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/creator-balances"
	gem_balances_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-balances"
	gem_packs_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var gemBalancesConf *config.Config
var gemBalancesDB *config.DB
var gemBalancesRepo gem_balances_repo.GemBalancesRepository
var gemPacksRepo gem_packs_repo.GemPacksRepository
var creatorBalancesRepo creator_balances_repo.CreatorBalancesRepository
var gemBalancesSvc *gemBalancesService

func TestMain(m *testing.M) {
	gemBalancesConf = config.CreateConfig()
	logger.InitLogger(false)
	var err error
	gemBalancesDB, err = config.NewDB(gemBalancesConf)
	if err != nil {
		panic(err)
	}
	gemBalancesRepo = gem_balances_repo.NewGemBalancesRepository(gemBalancesConf, gemBalancesDB)
	gemPacksRepo = gem_packs_repo.NewGemPacksRepository(gemBalancesConf, gemBalancesDB)
	creatorBalancesRepo = creator_balances_repo.NewCreatorBalancesRepository(gemBalancesConf, gemBalancesDB)
	gemBalancesSvc = &gemBalancesService{
		conf:                gemBalancesConf,
		gemBalancesRepo:     gemBalancesRepo,
		packsRepo:           gemPacksRepo,
		creatorBalancesRepo: creatorBalancesRepo,
	}

	clearGemBalancesTables()
	code := m.Run()
	clearGemBalancesTables()
	os.Exit(code)
}

func clearGemBalancesTables() {
	_ = gemBalancesDB.Conn.Exec("TRUNCATE TABLE gem_balances, processed_stripe_webhook_events RESTART IDENTITY CASCADE;")
}

func TestGemBalances_GetAll(t *testing.T) {
	clearGemBalancesTables()
	userID := uuid.New()
	_ = gemBalancesRepo.CreateGemBalance(userID)

	balances, err := gemBalancesSvc.GetAllGemBalances()
	assert.NoError(t, err)
	assert.Len(t, balances, 1)
}

func TestGemBalances_GetByUserId(t *testing.T) {
	clearGemBalancesTables()
	userID := uuid.New()
	_ = gemBalancesRepo.CreateGemBalance(userID)

	balance, err := gemBalancesSvc.GetGemBalanceByUserId(userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), balance.Gems)
}

func TestGemBalances_Create(t *testing.T) {
	clearGemBalancesTables()
	userID := uuid.New()

	err := gemBalancesSvc.CreateGemBalance(userID)
	assert.NoError(t, err)

	balance, err := gemBalancesRepo.GetGemBalanceByUserId(userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), balance.Gems)
}

func TestGemBalances_Update(t *testing.T) {
	clearGemBalancesTables()
	userID := uuid.New()

	err := gemBalancesSvc.UpdateGemBalance(userID, 42)
	assert.NoError(t, err)

	balance, err := gemBalancesRepo.GetGemBalanceByUserId(userID)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), balance.Gems)
}
