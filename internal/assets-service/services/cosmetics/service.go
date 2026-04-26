package cosmetics

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type CosmeticsService interface {
	// GetCategoriesList retrieves a list of sprite categories.
	GetCategoriesList() ([]*models.CosmeticCategory, error)

	// GetCosmeticsListByCategory retrieves a list of cosmetics for a given category.
	GetCosmeticsListByCategory(category uuid.UUID, worldId *uuid.UUID, playerId *uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error)

	// GetCosmeticById handles the retrieval of a cosmetic by its ID.
	GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error)

	// GetCosmeticsListByWorld retrieves a list of cosmetics for a given world.
	GetCosmeticsListByWorld(worldId uuid.UUID, offset int, limit int) ([]*models.Cosmetic, int64, error)

	// UploadCosmeticData handles the upload of cosmetic file.
	UploadCosmeticData(category uuid.UUID, worldId uuid.UUID, price float64, cosmeticData multipart.File, ext string, userId uuid.UUID) (*models.Cosmetic, error)

	// UploadCosmeticByID links an existing cosmetic sprite to another category.
	UploadCosmeticByID(categoryId uuid.UUID, worldId uuid.UUID, price float64, spriteId uuid.UUID, userId uuid.UUID) (*models.Cosmetic, error)

	// DeleteCosmetic handles the deletion of a cosmetic by its ID.
	DeleteCosmetic(cosmeticId uuid.UUID) error

	// AddCategory handles the addition of a new cosmetic category.
	AddCategory(category string) (*models.CosmeticCategory, error)

	// PurchaseCosmeticForUserInternal handles the purchase of a cosmetic for a user.
	PurchaseCosmeticForUserInternal(userId uuid.UUID, cosmeticId uuid.UUID) error
}
