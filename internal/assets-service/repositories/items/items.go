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
	return isr.db.Conn.Where("id = ?", item.Id).First(item).Error
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

func (isr *itemRepository) GetItemsListByWorld(worldid uuid.UUID) ([]*models.Item, error) {
	var items []*models.Item

	if err := isr.db.Conn.Where("world_id = ?", worldid).Find(&items).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewWorldNotFound("world not found")
		}
		return nil, err
	}

	if err := isr.db.Conn.
		Where("world_id = ?", worldid).
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
