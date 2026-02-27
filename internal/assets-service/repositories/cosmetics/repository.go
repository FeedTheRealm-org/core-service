package cosmetics

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// CosmeticsRepository defines the interface for cosmetics-related database operations.
type CosmeticsRepository interface {
	GetCategoriesList() ([]*models.CosmeticCategory, error)

	GetCosmeticsListByCategory(category uuid.UUID) ([]*models.Cosmetic, error)

	GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error)

	AddCategory(category string) (*models.CosmeticCategory, error)

	GetCategoryById(categoryId uuid.UUID) (*models.CosmeticCategory, error)

	CreateCosmetic(category uuid.UUID, cosmetic *models.Cosmetic) error
}
