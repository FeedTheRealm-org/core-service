package cosmetics

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/google/uuid"
)

type cosmeticsRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewCosmeticsRepository creates a new instance of CosmeticsRepository.
func NewCosmeticsRepository(conf *config.Config, db *config.DB) CosmeticsRepository {
	return &cosmeticsRepository{
		conf: conf,
		db:   db,
	}
}

func (cr *cosmeticsRepository) GetCategoriesList() ([]*models.CosmeticCategory, error) {
	var categories []*models.CosmeticCategory
	if err := cr.db.Conn.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (cr *cosmeticsRepository) GetCosmeticsListByCategory(category uuid.UUID) ([]*models.Cosmetic, error) {
	var cosmetics []*models.Cosmetic
	if err := cr.db.Conn.Where("category_id = ?", category).
		Find(&cosmetics).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewCategoryNotFound("category not found")
		}
		return nil, err
	}
	return cosmetics, nil
}

func (cr *cosmeticsRepository) GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error) {
	var cosmetic models.Cosmetic
	if err := cr.db.Conn.First(&cosmetic, "id = ?", cosmeticId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewCosmeticNotFound("cosmetic not found")
		}
		return nil, err
	}
	return &cosmetic, nil
}

func (cr *cosmeticsRepository) AddCategory(categoryName string) (*models.CosmeticCategory, error) {
	category := &models.CosmeticCategory{
		Name: categoryName,
	}

	if err := cr.db.Conn.Create(category).Error; err != nil {
		if errors.IsDuplicateEntryError(err) {
			return nil, assets_errors.NewCategoryConflict(err.Error())
		}
		return nil, err
	}

	return category, nil
}

func (cr *cosmeticsRepository) GetCategoryById(categoryId uuid.UUID) (*models.CosmeticCategory, error) {
	var category models.CosmeticCategory
	if err := cr.db.Conn.First(&category, "id = ?", categoryId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewCategoryNotFound("category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (cr *cosmeticsRepository) CreateCosmetic(categoryId uuid.UUID, cosmetic *models.Cosmetic) error {
	var category models.CosmeticCategory
	if err := cr.db.Conn.First(&category, "id = ?", categoryId).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return assets_errors.NewCategoryNotFound("category not found")
		}
		return err
	}

	cosmetic.CategoryID = category.Id
	if err := cr.db.Conn.Create(cosmetic).Error; err != nil {
		return err
	}

	return nil
}
