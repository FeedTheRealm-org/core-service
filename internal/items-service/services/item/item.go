package item

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/repositories/item"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type itemService struct {
	conf           *config.Config
	itemRepository item.ItemRepository
}

// NewItemService creates a new instance of ItemService.
func NewItemService(conf *config.Config, itemRepository item.ItemRepository) ItemService {
	return &itemService{
		conf:           conf,
		itemRepository: itemRepository,
	}
}

func (is *itemService) CreateItem(newItem *models.Item) error {
	if err := is.itemRepository.CreateItem(newItem); err != nil {
		return err
	}
	logger.Logger.Infof("Item created: %s (ID: %s)", newItem.Name, newItem.Id)
	return nil
}

func (is *itemService) CreateItems(items []models.Item) error {
	if err := is.itemRepository.CreateItems(items); err != nil {
		return err
	}
	logger.Logger.Infof("Batch created %d items", len(items))
	return nil
}

func (is *itemService) GetItemById(id uuid.UUID) (*models.Item, error) {
	item, err := is.itemRepository.GetItemById(id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (is *itemService) GetAllItems() ([]models.Item, error) {
	items, err := is.itemRepository.GetAllItems()
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (is *itemService) DeleteItem(id uuid.UUID) error {
	if err := is.itemRepository.DeleteItem(id); err != nil {
		return err
	}
	logger.Logger.Infof("Item deleted: %s", id)
	return nil
}

func (is *itemService) ClearAllItems() error {
	return is.itemRepository.DeleteAll()
}

func (is *itemService) UpdateItemSprite(id uuid.UUID, spriteId uuid.UUID) error {
	item, err := is.itemRepository.GetItemById(id)
	if err != nil {
		return err
	}

	item.SpriteId = spriteId
	return is.itemRepository.UpdateItem(item)
}
