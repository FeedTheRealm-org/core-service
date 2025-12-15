package item

import (
	"fmt"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	item_errors "github.com/FeedTheRealm-org/core-service/internal/items-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/google/uuid"
)

type itemRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewItemRepository creates a new instance of ItemRepository.
func NewItemRepository(conf *config.Config, db *config.DB) ItemRepository {
	return &itemRepository{
		conf: conf,
		db:   db,
	}
}

func (ir *itemRepository) CreateItem(item *models.Item) error {
	if err := ir.db.Conn.Create(item).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return item_errors.NewItemAlreadyExists(item.Name)
		}
		return err
	}
	return nil
}

func (ir *itemRepository) CreateItems(items []models.Item) error {
	if err := ir.db.Conn.Create(&items).Error; err != nil {
		return err
	}
	return nil
}

func (ir *itemRepository) GetItemById(id uuid.UUID) (*models.Item, error) {
	var item models.Item
	if err := ir.db.Conn.Where("id = ?", id).First(&item).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, item_errors.NewItemNotFound(err.Error())
		}
		return nil, err
	}
	return &item, nil
}

func (ir *itemRepository) GetAllItems() ([]models.Item, error) {
	var items []models.Item
	if err := ir.db.Conn.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (ir *itemRepository) GetItemsByCategory(categoryId uuid.UUID) ([]models.Item, error) {
	var items []models.Item
	if err := ir.db.Conn.Where("category_id = ?", categoryId).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (ir *itemRepository) UpdateItem(item *models.Item) error {
	if err := ir.db.Conn.Save(item).Error; err != nil {
		return err
	}
	return nil
}

func (ir *itemRepository) DeleteItem(id uuid.UUID) error {
	if err := ir.db.Conn.Delete(&models.Item{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (ir *itemRepository) DeleteAll() error {
	if os.Getenv("ALLOW_DB_RESET") != "true" {
		return fmt.Errorf("forbidden: database reset not allowed")
	}
	return ir.db.Conn.Exec("DELETE FROM items").Error
}
