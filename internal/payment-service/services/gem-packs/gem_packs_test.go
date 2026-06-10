package gem_packs

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	gem_packs_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var gemPacksConf *config.Config
var gemPacksDB *config.DB
var gemPacksRepo gem_packs_repo.GemPacksRepository
var gemPacksSvc *gemPacksService

func TestMain(m *testing.M) {
	gemPacksConf = config.CreateConfig()
	logger.InitLogger(false)
	var err error
	gemPacksDB, err = config.NewDB(gemPacksConf)
	if err != nil {
		panic(err)
	}
	gemPacksRepo = gem_packs_repo.NewGemPacksRepository(gemPacksConf, gemPacksDB)
	gemPacksSvc = &gemPacksService{conf: gemPacksConf, repo: gemPacksRepo}

	clearGemPacksTables()
	code := m.Run()
	clearGemPacksTables()
	os.Exit(code)
}

func clearGemPacksTables() {
	_ = gemPacksDB.Conn.Exec("TRUNCATE TABLE gem_packs RESTART IDENTITY CASCADE;")
}

func TestGemPacks_CreateAndGet(t *testing.T) {
	clearGemPacksTables()

	created, err := gemPacksSvc.CreateGemPack("Starter", 100, decimal.NewFromFloat(1.25))
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.Id)
	assert.Equal(t, "Starter", created.Name)

	fetched, err := gemPacksSvc.GetGemPackById(created.Id)
	assert.NoError(t, err)
	assert.Equal(t, created.Id, fetched.Id)
}

func TestGemPacks_UpdatePartialFields(t *testing.T) {
	clearGemPacksTables()

	pack, err := gemPacksSvc.CreateGemPack("Starter", 100, decimal.NewFromFloat(2))
	assert.NoError(t, err)

	updated, err := gemPacksSvc.UpdateGemPack(pack.Id, "", 0, decimal.Zero)
	assert.NoError(t, err)
	assert.Equal(t, "Starter", updated.Name)
	assert.Equal(t, int64(100), updated.Gems)
	assert.True(t, updated.Price.Equal(decimal.NewFromFloat(2)))

	updated, err = gemPacksSvc.UpdateGemPack(pack.Id, "Pro", 500, decimal.NewFromFloat(4.5))
	assert.NoError(t, err)
	assert.Equal(t, "Pro", updated.Name)
	assert.Equal(t, int64(500), updated.Gems)
	assert.True(t, updated.Price.Equal(decimal.NewFromFloat(4.5)))
}

func TestGemPacks_Seed_NoConfig(t *testing.T) {
	clearGemPacksTables()
	_, err := gemPacksSvc.CreateGemPack("Existing", 50, decimal.NewFromFloat(0.5))
	assert.NoError(t, err)

	originalPacks := gemPacksConf.Stripe.GemPacks
	gemPacksConf.Stripe.GemPacks = nil
	defer func() {
		gemPacksConf.Stripe.GemPacks = originalPacks
	}()

	err = gemPacksSvc.seedPacksData()
	assert.NoError(t, err)
	packs, err := gemPacksSvc.GetAllGemPacks()
	assert.NoError(t, err)
	assert.Len(t, packs, 1)
}

func TestGemPacks_Seed_ReplacesExisting(t *testing.T) {
	clearGemPacksTables()
	_, err := gemPacksSvc.CreateGemPack("Old", 10, decimal.NewFromFloat(0.5))
	assert.NoError(t, err)

	originalPacks := gemPacksConf.Stripe.GemPacks
	gemPacksConf.Stripe.GemPacks = []config.StripeItem{
		{Name: "Starter", Amount: 100, Price: 1.5},
		{Name: "Pro", Amount: 500, Price: 4.5},
	}
	defer func() {
		gemPacksConf.Stripe.GemPacks = originalPacks
	}()

	err = gemPacksSvc.seedPacksData()
	assert.NoError(t, err)
	packs, err := gemPacksSvc.GetAllGemPacks()
	assert.NoError(t, err)
	assert.Len(t, packs, 2)
}
