package cosmetics_test

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	cosmeticsrepo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/cosmetics"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var cosmeticsConf *config.Config
var cosmeticsDB *config.DB
var cosmeticsRepo cosmeticsrepo.CosmeticsRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)

	cosmeticsConf = config.CreateConfig()
	var err error
	cosmeticsDB, err = config.NewDB(cosmeticsConf)
	if err != nil {
		panic(err)
	}
	cosmeticsRepo = cosmeticsrepo.NewCosmeticsRepository(cosmeticsConf, cosmeticsDB)

	clearCosmeticsTables()
	code := m.Run()
	clearCosmeticsTables()
	os.Exit(code)
}

func clearCosmeticsTables() {
	_ = cosmeticsDB.Conn.Exec("TRUNCATE TABLE purchases, cosmetics, cosmetics_categories RESTART IDENTITY CASCADE;")
}

func TestCosmeticsRepository_AddCategoryAndGet(t *testing.T) {
	clearCosmeticsTables()

	category, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)
	assert.NotNil(t, category)

	stored, err := cosmeticsRepo.GetCategoryById(category.Id)
	assert.NoError(t, err)
	assert.Equal(t, "hats", stored.Name)
}

func TestCosmeticsRepository_AddCategory_Duplicate(t *testing.T) {
	clearCosmeticsTables()

	_, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)
	_, err = cosmeticsRepo.AddCategory("hats")
	assert.Error(t, err)

	var conflict *assets_errors.CategoryConflict
	assert.True(t, errors.As(err, &conflict))
}

func TestCosmeticsRepository_CreateAndGetCosmetic(t *testing.T) {
	clearCosmeticsTables()

	category, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)

	cosmetic := &models.Cosmetic{Url: "/hats/one.png"}
	worldId := uuid.New()
	createdBy := uuid.New()
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, worldId, 15, cosmetic, createdBy))

	stored, err := cosmeticsRepo.GetCosmeticById(cosmetic.Id)
	assert.NoError(t, err)
	assert.Equal(t, cosmetic.Url, stored.Url)
}

func TestCosmeticsRepository_GetCosmeticById_NotFound(t *testing.T) {
	clearCosmeticsTables()

	cosmetic, err := cosmeticsRepo.GetCosmeticById(uuid.New())
	assert.Error(t, err)
	assert.Nil(t, cosmetic)

	var notFound *assets_errors.CosmeticNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestCosmeticsRepository_AddPurchaseForUserId(t *testing.T) {
	clearCosmeticsTables()

	category, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)

	cosmetic := &models.Cosmetic{Url: "/hats/one.png"}
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, uuid.New(), 10, cosmetic, uuid.New()))

	userId := uuid.New()
	assert.NoError(t, cosmeticsRepo.AddPurchaseForUserId(cosmetic.Id, userId))

	err = cosmeticsRepo.AddPurchaseForUserId(cosmetic.Id, userId)
	assert.Error(t, err)

	var conflict *assets_errors.CosmeticsWasPurchasedBefore
	assert.True(t, errors.As(err, &conflict))
}

func TestCosmeticsRepository_GetCosmeticsListByCategory_Branches(t *testing.T) {
	clearCosmeticsTables()

	category, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)

	worldId := uuid.New()
	playerId := uuid.New()
	defaultCosmetic := &models.Cosmetic{Url: "/hats/default.png"}
	worldCosmetic := &models.Cosmetic{Url: "/hats/world.png"}
	playerCosmetic := &models.Cosmetic{Url: "/hats/player.png"}

	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, uuid.Nil, 5, defaultCosmetic, uuid.New()))
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, worldId, 7, worldCosmetic, uuid.New()))
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, worldId, 9, playerCosmetic, uuid.New()))
	assert.NoError(t, cosmeticsRepo.AddPurchaseForUserId(playerCosmetic.Id, playerId))

	list, total, err := cosmeticsRepo.GetCosmeticsListByCategory(category.Id, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	list, total, err = cosmeticsRepo.GetCosmeticsListByCategory(category.Id, &worldId, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)

	list, total, err = cosmeticsRepo.GetCosmeticsListByCategory(category.Id, nil, &playerId, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	list, total, err = cosmeticsRepo.GetCosmeticsListByCategory(category.Id, &worldId, &playerId, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestCosmeticsRepository_UpdateAndDelete(t *testing.T) {
	clearCosmeticsTables()

	category, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)

	cosmetic := &models.Cosmetic{Url: "/hats/one.png"}
	worldId := uuid.New()
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, worldId, 10, cosmetic, uuid.New()))

	assert.NoError(t, cosmeticsRepo.UpdateCosmetic(cosmetic.Id, 20, ""))
	updated, err := cosmeticsRepo.GetCosmeticById(cosmetic.Id)
	assert.NoError(t, err)
	assert.Equal(t, int64(20), updated.Price)

	assert.NoError(t, cosmeticsRepo.UpdateCosmetic(cosmetic.Id, 30, "/hats/two.png"))
	updated, err = cosmeticsRepo.GetCosmeticById(cosmetic.Id)
	assert.NoError(t, err)
	assert.Equal(t, "/hats/two.png", updated.Url)

	assert.NoError(t, cosmeticsRepo.DeleteCosmetic(cosmetic.Id))
}

func TestCosmeticsRepository_GetCategoriesAndEconomySummary(t *testing.T) {
	clearCosmeticsTables()

	cat1, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)
	cat2, err := cosmeticsRepo.AddCategory("shoes")
	assert.NoError(t, err)

	list, err := cosmeticsRepo.GetCategoriesList()
	assert.NoError(t, err)
	assert.Len(t, list, 2)

	defaultCosmetic := &models.Cosmetic{Url: "/hats/default.png"}
	worldCosmetic := &models.Cosmetic{Url: "/shoes/world.png"}
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(cat1.Id, uuid.Nil, 5, defaultCosmetic, uuid.New()))
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(cat2.Id, uuid.New(), 10, worldCosmetic, uuid.New()))

	summary, err := cosmeticsRepo.GetEconomySummary()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), summary.DefaultCosmetics)
	assert.Equal(t, int64(1), summary.UserCreatedCosmetics)
}

func TestCosmeticsRepository_GetCosmeticsListByWorld(t *testing.T) {
	clearCosmeticsTables()

	category, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)

	worldId := uuid.New()
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, worldId, 10, &models.Cosmetic{Url: "/hats/world.png"}, uuid.New()))
	assert.NoError(t, cosmeticsRepo.CreateCosmetic(category.Id, uuid.New(), 10, &models.Cosmetic{Url: "/hats/other.png"}, uuid.New()))

	list, total, err := cosmeticsRepo.GetCosmeticsListByWorld(worldId, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestCosmeticsRepository_GetCosmeticByUrlCategoryAndWorld_NotFound(t *testing.T) {
	clearCosmeticsTables()

	category, err := cosmeticsRepo.AddCategory("hats")
	assert.NoError(t, err)

	cosmetic, err := cosmeticsRepo.GetCosmeticByUrlCategoryAndWorld("/missing.png", category.Id, uuid.New())
	assert.Error(t, err)
	assert.Nil(t, cosmetic)

	var notFound *assets_errors.CosmeticNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestCosmeticsRepository_CreateCosmetic_CategoryNotFound(t *testing.T) {
	clearCosmeticsTables()

	err := cosmeticsRepo.CreateCosmetic(uuid.New(), uuid.New(), 10, &models.Cosmetic{Url: "/missing.png"}, uuid.New())
	assert.Error(t, err)

	var notFound *assets_errors.CategoryNotFound
	assert.True(t, errors.As(err, &notFound))
}
