package cosmetics

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// CosmeticsRepository defines the interface for cosmetics-related database operations.
type CosmeticsRepository interface {
	GetCategoriesList() ([]*models.CosmeticCategory, error)

	GetCosmeticsListByCategory(category uuid.UUID, worldId *uuid.UUID, playerId *uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error)

	GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error)

	AddCategory(category string) (*models.CosmeticCategory, error)

	AddPurchaseForUserId(cosmeticId uuid.UUID, userId uuid.UUID) error

	GetCategoryById(categoryId uuid.UUID) (*models.CosmeticCategory, error)

	GetCosmeticsListByWorld(worldId uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error)

	CreateCosmetic(category uuid.UUID, worldId uuid.UUID, price float64, cosmetic *models.Cosmetic, userId uuid.UUID) error

	DeleteCosmetic(cosmeticId uuid.UUID) error
}
