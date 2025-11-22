package item

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	item_errors "github.com/FeedTheRealm-org/core-service/internal/items-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/google/uuid"
)

type itemSpriteRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewItemSpriteRepository creates a new instance of ItemSpriteRepository.
func NewItemSpriteRepository(conf *config.Config, db *config.DB) ItemSpriteRepository {
	return &itemSpriteRepository{
		conf: conf,
		db:   db,
	}
}

func (isr *itemSpriteRepository) CreateSprite(sprite *models.ItemSprite) error {
	if err := isr.db.Conn.Create(sprite).Error; err != nil {
		return err
	}
	return nil
}

func (isr *itemSpriteRepository) GetSpriteById(id uuid.UUID) (*models.ItemSprite, error) {
	var sprite models.ItemSprite
	if err := isr.db.Conn.Where("id = ?", id).First(&sprite).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, item_errors.NewItemSpriteNotFound(err.Error())
		}
		return nil, err
	}
	return &sprite, nil
}

func (isr *itemSpriteRepository) GetSpritesByCategory(category string) ([]models.ItemSprite, error) {
	var sprites []models.ItemSprite
	if err := isr.db.Conn.Where("category = ?", category).Find(&sprites).Error; err != nil {
		return nil, err
	}
	return sprites, nil
}

func (isr *itemSpriteRepository) DeleteSprite(id uuid.UUID) error {
	if err := isr.db.Conn.Delete(&models.ItemSprite{}, id).Error; err != nil {
		return err
	}
	return nil
}
