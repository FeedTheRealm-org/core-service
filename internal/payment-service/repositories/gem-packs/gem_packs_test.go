package gem_packs

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	core_errors "github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var gemPacksConf *config.Config
var gemPacksDB *config.DB
var gemPacksRepo GemPacksRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)
	gemPacksConf = config.CreateConfig()
	var err error
	gemPacksDB, err = config.NewDB(gemPacksConf)
	if err != nil {
		panic(err)
	}
	gemPacksRepo = NewGemPacksRepository(gemPacksConf, gemPacksDB)

	clearGemPacksTables()
	code := m.Run()
	clearGemPacksTables()
	os.Exit(code)
}

func clearGemPacksTables() {
	_ = gemPacksDB.Conn.Exec("TRUNCATE TABLE gem_packs RESTART IDENTITY CASCADE;")
}

func TestGemPacksRepository_CreateGetUpdateDelete(t *testing.T) {
	clearGemPacksTables()

	pack := &models.GemPack{
		Id:    uuid.New(),
		Name:  "Starter",
		Gems:  100,
		Price: decimal.NewFromFloat(1.25),
	}
	created, err := gemPacksRepo.CreateGemPack(pack)
	assert.NoError(t, err)
	assert.Equal(t, pack.Id, created.Id)

	all, err := gemPacksRepo.GetAllGemPacks()
	assert.NoError(t, err)
	assert.Len(t, all, 1)

	found, err := gemPacksRepo.GetGemPackById(pack.Id)
	assert.NoError(t, err)
	assert.Equal(t, "Starter", found.Name)

	updatedPack := &models.GemPack{Name: "Pro", Gems: 200, Price: decimal.NewFromFloat(2.5)}
	assert.NoError(t, gemPacksRepo.UpdateGemPack(pack.Id, updatedPack))

	updated, err := gemPacksRepo.GetGemPackById(pack.Id)
	assert.NoError(t, err)
	assert.Equal(t, "Pro", updated.Name)
	assert.Equal(t, int64(200), updated.Gems)

	assert.NoError(t, gemPacksRepo.DeleteGemPack(pack.Id))
	_, err = gemPacksRepo.GetGemPackById(pack.Id)
	assert.Error(t, err)
	var notFound *core_errors.HttpError
	assert.True(t, errors.As(err, &notFound))
	assert.Equal(t, 404, notFound.Status)
}
