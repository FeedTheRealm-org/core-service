package materials

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// MaterialsService defines the interface for material-related business logic operations.
type MaterialsService interface {
	// UploadMaterial uploads or overwrites a single material file with the provided ID for a given world and saves its metadata.
	UploadMaterial(worldID uuid.UUID, id uuid.UUID, materialType models.MaterialType, name string, file *multipart.FileHeader, userId uuid.UUID) (*models.Material, error)

	// GetMaterialByID retrieves a material by its ID.
	GetMaterialByID(id uuid.UUID) (*models.Material, error)

	// GetMaterialsListByWorld retrieves a list of materials for a specific world, including default materials.
	GetMaterialsListByWorld(worldID uuid.UUID, offset int, limit int) ([]*models.Material, error)

	// GetMaterialsListByWorldAndType retrieves a list of materials for a specific world and material type, including default materials.
	GetMaterialsListByWorldAndType(worldID uuid.UUID, materialType models.MaterialType, offset int, limit int) ([]*models.Material, error)

	// DeleteMaterial deletes a material by its ID.
	DeleteMaterial(id uuid.UUID) error
}
