package items

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type itemSpritesRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewItemSpritesRepository creates a new instance of ItemSpritesRepository.
func NewItemSpritesRepository(conf *config.Config, db *config.DB) ItemSpritesRepository {
	return &itemSpritesRepository{
		conf: conf,
		db:   db,
	}
}

func (isr *itemSpritesRepository) UpsertSprite(sprite *models.ItemSprite) error {
	if err := isr.db.Conn.
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"url", "updated_at"}),
			},
		).Create(sprite).Error; err != nil {
		return err
	}
	return nil
}

func (isr *itemSpritesRepository) GetSpriteById(id uuid.UUID) (*models.ItemSprite, error) {
	var sprite models.ItemSprite
	if err := isr.db.Conn.Where("id = ?", id).First(&sprite).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewItemSpriteNotFound("item sprite not found")
		}
		return nil, err
	}
	return &sprite, nil
}

func (isr *itemSpritesRepository) GetAllSprites() ([]models.ItemSprite, error) {
	var sprites []models.ItemSprite
	if err := isr.db.Conn.Find(&sprites).Error; err != nil {
		return nil, err
	}
	return sprites, nil
}

func (isr *itemSpritesRepository) DeleteSprite(id uuid.UUID) error {
	if err := isr.db.Conn.Delete(&models.ItemSprite{}, id).Error; err != nil {
		return err
	}
	return nil
}
