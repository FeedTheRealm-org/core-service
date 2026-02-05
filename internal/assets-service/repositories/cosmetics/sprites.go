package cosmetics

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/google/uuid"
)

type spritesRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewSpritesRepository creates a new instance of SpritesRepository.
func NewSpritesRepository(conf *config.Config, db *config.DB) SpritesRepository {
	return &spritesRepository{
		conf: conf,
		db:   db,
	}
}

func (sr *spritesRepository) GetCategoriesList() ([]*models.Category, error) {
	var categories []*models.Category
	if err := sr.db.Conn.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (sr *spritesRepository) GetSpritesListByCategory(category uuid.UUID) ([]*models.Sprite, error) {
	var sprites []*models.Sprite
	if err := sr.db.Conn.Joins("JOIN sprite_categories sc ON sc.sprite_id = sprites.id").
		Where("sc.category_id = ?", category).
		Find(&sprites).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewCategoryNotFound("category not found")
		}
		return nil, err
	}
	return sprites, nil
}

func (sr *spritesRepository) GetSpriteById(spriteId uuid.UUID) (*models.Sprite, error) {
	var sprite models.Sprite
	if err := sr.db.Conn.First(&sprite, "id = ?", spriteId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewSpriteNotFound("sprite not found")
		}
		return nil, err
	}
	return &sprite, nil
}

func (sr *spritesRepository) AddCategory(categoryName string) (*models.Category, error) {
	category := &models.Category{
		Name: categoryName,
	}
	if err := sr.db.Conn.Create(category).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return nil, assets_errors.NewCategoryConflict(err.Error())
		}
		return nil, err
	}
	return category, nil
}

func (sr *spritesRepository) CreateSprite(categoryId uuid.UUID, sprite *models.Sprite) error {
	var category models.Category
	if err := sr.db.Conn.First(&category, "id = ?", categoryId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return assets_errors.NewCategoryNotFound("category not found")
		}
		return err
	}

	sprite.Categories = []models.Category{category}

	if err := sr.db.Conn.Create(sprite).Error; err != nil {
		return err
	}

	return nil
}
