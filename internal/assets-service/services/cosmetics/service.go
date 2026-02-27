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
	GetCosmeticsListByCategory(category uuid.UUID) ([]*models.Cosmetic, error)

	// GetCosmeticById handles the retrieval of a cosmetic by its ID.
	GetCosmeticById(cosmeticId uuid.UUID) (*models.Cosmetic, error)

	// UploadCosmeticData handles the upload of cosmetic file.
	UploadCosmeticData(category uuid.UUID, cosmeticData multipart.File, ext string) (*models.Cosmetic, error)

	// AddCategory handles the addition of a new cosmetic category.
	AddCategory(category string) (*models.CosmeticCategory, error)
}
