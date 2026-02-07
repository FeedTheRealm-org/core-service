package items

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
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

func (isr *itemRepository) UpsertItem(item *models.Item) error {
	if err := isr.db.Conn.
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"url", "updated_at"}),
			},
		).Create(item).Error; err != nil {
		return err
	}
	return nil
}

func (isr *itemRepository) GetItemById(id uuid.UUID) (*models.Item, error) {
	var item models.Item
	if err := isr.db.Conn.Where("id = ?", id).First(&item).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewItemSpriteNotFound("item sprite not found")
		}
		return nil, err
	}
	return &item, nil
}

func (isr *itemRepository) GetAllItems() ([]*models.Item, error) {
	var items []*models.Item
	if err := isr.db.Conn.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (isr *itemRepository) GetItemsListByCategory(worldid uuid.UUID, categoryId uuid.UUID) ([]*models.Item, error) {
	var items []*models.Item

	if err := isr.db.Conn.Where("world_id = ?", worldid).Find(&items).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewWorldNotFound("category not found")
		}
		return nil, err
	}

	if err := isr.db.Conn.Where("category_id = ?", categoryId).Find(&items).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewCategoryNotFound("category not found")
		}
		return nil, err
	}

	if err := isr.db.Conn.
		Where("world_id = ? AND category_id = ?", worldid, categoryId).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (isr *itemRepository) DeleteSprite(id uuid.UUID) error {
	if err := isr.db.Conn.Delete(&models.Item{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (isr *itemRepository) AddCategory(name string) (*models.ItemCategory, error) {
	category := &models.ItemCategory{
		Name: name,
	}

	if err := isr.db.Conn.Create(category).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return nil, assets_errors.NewCategoryConflict(err.Error())
		}
		return nil, err
	}

	return category, nil
}
