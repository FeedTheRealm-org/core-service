package items_test

import (
	"errors"
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	itemsrepo "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/items"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var itemsConf *config.Config
var itemsDB *config.DB
var itemsRepo itemsrepo.ItemRepository

func TestMain(m *testing.M) {
	logger.InitLogger(false)

	itemsConf = config.CreateConfig()
	var err error
	itemsDB, err = config.NewDB(itemsConf)
	if err != nil {
		panic(err)
	}
	itemsRepo = itemsrepo.NewItemRepository(itemsConf, itemsDB)

	clearItemsTable()
	code := m.Run()
	clearItemsTable()
	os.Exit(code)
}

func clearItemsTable() {
	_ = itemsDB.Conn.Exec("TRUNCATE TABLE items RESTART IDENTITY CASCADE;")
}

func TestItemRepository_UpsertAndGet(t *testing.T) {
	clearItemsTable()

	item := &models.Item{
		Id:        uuid.New(),
		WorldID:   uuid.New(),
		Url:       "/worlds/a/items/b.png",
		CreatedBy: uuid.New(),
	}

	err := itemsRepo.UpsertItem(item)
	assert.NoError(t, err)

	stored, err := itemsRepo.GetItemById(item.Id)
	assert.NoError(t, err)
	assert.NotNil(t, stored)
	assert.Equal(t, item.Url, stored.Url)
}

func TestItemRepository_GetItemById_NotFound(t *testing.T) {
	clearItemsTable()

	item, err := itemsRepo.GetItemById(uuid.New())
	assert.Error(t, err)
	assert.Nil(t, item)

	var notFound *assets_errors.ItemSpriteNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestItemRepository_GetAllAndDelete(t *testing.T) {
	clearItemsTable()

	item := &models.Item{
		Id:        uuid.New(),
		WorldID:   uuid.New(),
		Url:       "/worlds/a/items/b.png",
		CreatedBy: uuid.New(),
	}
	assert.NoError(t, itemsRepo.UpsertItem(item))

	items, err := itemsRepo.GetAllItems()
	assert.NoError(t, err)
	assert.Len(t, items, 1)

	assert.NoError(t, itemsRepo.DeleteSprite(item.Id))
	items, err = itemsRepo.GetAllItems()
	assert.NoError(t, err)
	assert.Len(t, items, 0)
}

func TestItemRepository_GetItemsListByWorld(t *testing.T) {
	clearItemsTable()

	worldID := uuid.New()
	otherWorld := uuid.New()
	assert.NoError(t, itemsRepo.UpsertItem(&models.Item{Id: uuid.New(), WorldID: worldID, Url: "/w/a.png", CreatedBy: uuid.New()}))
	assert.NoError(t, itemsRepo.UpsertItem(&models.Item{Id: uuid.New(), WorldID: otherWorld, Url: "/w/b.png", CreatedBy: uuid.New()}))

	items, err := itemsRepo.GetItemsListByWorld(worldID)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestItemRepository_UpsertItem_UpdatesExisting(t *testing.T) {
	clearItemsTable()

	itemID := uuid.New()
	worldID := uuid.New()
	assert.NoError(t, itemsRepo.UpsertItem(&models.Item{Id: itemID, WorldID: worldID, Url: "/v1.png", CreatedBy: uuid.New()}))

	assert.NoError(t, itemsRepo.UpsertItem(&models.Item{Id: itemID, WorldID: worldID, Url: "/v2.png", CreatedBy: uuid.New()}))
	stored, err := itemsRepo.GetItemById(itemID)
	assert.NoError(t, err)
	assert.Equal(t, "/v2.png", stored.Url)
}
