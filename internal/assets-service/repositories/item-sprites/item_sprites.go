package itemsprites

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
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

func (isr *itemSpritesRepository) CreateSprite(sprite *models.ItemSprite) error {
	if err := isr.db.Conn.Create(sprite).Error; err != nil {
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

func (isr *itemSpritesRepository) GetSpritesByCategory(categoryId uuid.UUID) ([]models.ItemSprite, error) {
	var sprites []models.ItemSprite
	if err := isr.db.Conn.Where("category_id = ?", categoryId).Find(&sprites).Error; err != nil {
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

// GetCategoryById reads from items-service table for validation
func (isr *itemSpritesRepository) GetCategoryById(id uuid.UUID) (*models.ItemCategory, error) {
	var category models.ItemCategory
	if err := isr.db.Conn.Where("id = ?", id).First(&category).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewItemCategoryNotFound(id.String())
		}
		return nil, err
	}
	return &category, nil
}

// GetAllCategories reads all categories from items-service table
func (isr *itemSpritesRepository) GetAllCategories() ([]models.ItemCategory, error) {
	var categories []models.ItemCategory
	if err := isr.db.Conn.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
