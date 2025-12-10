package item

import (
	"github.com/FeedTheRealm-org/core-service/config"
	item_errors "github.com/FeedTheRealm-org/core-service/internal/items-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/repositories/item"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type itemCategoryService struct {
	conf         *config.Config
	categoryRepo item.ItemCategoryRepository
}

// NewItemCategoryService creates a new instance of ItemCategoryService.
func NewItemCategoryService(conf *config.Config, categoryRepo item.ItemCategoryRepository) ItemCategoryService {
	return &itemCategoryService{
		conf:         conf,
		categoryRepo: categoryRepo,
	}
}

func (ics *itemCategoryService) CreateCategory(name string) (*models.ItemCategory, error) {
	category := &models.ItemCategory{
		Name: name,
	}

	if err := ics.categoryRepo.CreateCategory(category); err != nil {
		return nil, err
	}

	logger.Logger.Infof("Item category created: %s (ID: %s)", category.Name, category.Id)
	return category, nil
}

func (ics *itemCategoryService) GetCategoryById(id uuid.UUID) (*models.ItemCategory, error) {
	category, err := ics.categoryRepo.GetCategoryById(id)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (ics *itemCategoryService) GetAllCategories() ([]models.ItemCategory, error) {
	return ics.categoryRepo.GetAllCategories()
}

func (ics *itemCategoryService) DeleteCategory(id uuid.UUID) error {
	// First check if category exists
	category, err := ics.categoryRepo.GetCategoryById(id)
	if err != nil {
		return err
	}

	// Check if category is in use by items
	itemCount, err := ics.categoryRepo.CountItemsUsingCategory(id)
	if err != nil {
		return err
	}

	// Check if category is in use by sprites
	spriteCount, err := ics.categoryRepo.CountSpritesUsingCategory(id)
	if err != nil {
		return err
	}

	// If category is in use, return error
	if itemCount > 0 || spriteCount > 0 {
		return item_errors.NewItemCategoryInUse(category.Name, itemCount, spriteCount)
	}

	// Delete category
	if err := ics.categoryRepo.DeleteCategory(id); err != nil {
		return err
	}

	logger.Logger.Infof("Item category deleted: %s (ID: %s)", category.Name, id)
	return nil
}

func (ics *itemCategoryService) ClearAllCategories() error {
	return ics.categoryRepo.DeleteAll()
}
