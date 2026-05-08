package materials

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type MaterialsRepository interface {
	// GetMaterialsListByWorld retrieves a list of materials for a specific world, including default materials.
	GetMaterialsListByWorld(worldID uuid.UUID, offset int, limit int) ([]*models.Material, error)

	// GetMaterialsListByWorldAndType retrieves a list of materials for a specific world and material type, including default materials.
	GetMaterialsListByWorldAndType(worldID uuid.UUID, materialType models.MaterialType, offset int, limit int) ([]*models.Material, error)

	// GetMaterialByID retrieves a single material by its ID.
	GetMaterialByID(materialID uuid.UUID) (*models.Material, error)

	// UpsertMaterial creates a new material or updates an existing one.
	UpsertMaterial(material *models.Material) error

	// DeleteMaterial removes a material from the database.
	DeleteMaterial(material *models.Material) error
}
