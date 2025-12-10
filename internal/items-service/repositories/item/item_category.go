package item

import (
	"fmt"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	item_errors "github.com/FeedTheRealm-org/core-service/internal/items-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type itemCategoryRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewItemCategoryRepository creates a new instance of ItemCategoryRepository.
func NewItemCategoryRepository(conf *config.Config, db *config.DB) ItemCategoryRepository {
	return &itemCategoryRepository{
		conf: conf,
		db:   db,
	}
}

func (icr *itemCategoryRepository) CreateCategory(category *models.ItemCategory) error {
	if err := icr.db.Conn.Create(category).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return item_errors.NewItemCategoryConflict("category name already exists")
		}
		return err
	}
	return nil
}

func (icr *itemCategoryRepository) GetCategoryById(id uuid.UUID) (*models.ItemCategory, error) {
	var category models.ItemCategory
	if err := icr.db.Conn.Where("id = ?", id).First(&category).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, item_errors.NewItemCategoryNotFound(id.String())
		}
		return nil, err
	}
	return &category, nil
}

func (icr *itemCategoryRepository) GetAllCategories() ([]models.ItemCategory, error) {
	var categories []models.ItemCategory
	if err := icr.db.Conn.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (icr *itemCategoryRepository) DeleteCategory(id uuid.UUID) error {
	if err := icr.db.Conn.Delete(&models.ItemCategory{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (icr *itemCategoryRepository) CountItemsUsingCategory(categoryId uuid.UUID) (int64, error) {
	var count int64
	if err := icr.db.Conn.Model(&models.Item{}).Where("category_id = ?", categoryId).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (icr *itemCategoryRepository) CountSpritesUsingCategory(categoryId uuid.UUID) (int64, error) {
	var count int64
	// Query the item_sprites table directly (managed by assets-service, but in same DB)
	if err := icr.db.Conn.Table("item_sprites").Where("category_id = ?", categoryId).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (icr *itemCategoryRepository) DeleteAll() error {
	if os.Getenv("ALLOW_DB_RESET") != "true" {
		return fmt.Errorf("forbidden: database reset not allowed")
	}
	// Delete sprites that reference these categories first to avoid FK violations.
	return icr.db.Conn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM item_sprites").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM item_categories").Error; err != nil {
			return err
		}
		return nil
	})
}
