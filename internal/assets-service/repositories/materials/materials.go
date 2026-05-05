package materials

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assets_errors "github.com/FeedTheRealm-org/core-service/internal/assets-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type materialsRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewMaterialsRepository creates a new instance of MaterialsRepository.
func NewMaterialsRepository(conf *config.Config, db *config.DB) MaterialsRepository {
	return &materialsRepository{
		conf: conf,
		db:   db,
	}
}

func (mr *materialsRepository) UpsertMaterial(material *models.Material) error {
	if err := mr.db.Conn.
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"url", "updated_at"}),
			},
		).Create(material).Error; err != nil {
		return err
	}
	return mr.db.Conn.Where("id = ?", material.ID).First(material).Error
}

func (mr *materialsRepository) GetMaterialByID(id uuid.UUID) (*models.Material, error) {
	var material models.Material
	if err := mr.db.Conn.Where("id = ?", id).First(&material).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, assets_errors.NewMaterialNotFound("material not found")
		}
		return nil, err
	}
	return &material, nil
}

func (mr *materialsRepository) GetMaterialsListByWorld(worldID uuid.UUID, offset int, limit int) ([]*models.Material, error) {
	var materials []*models.Material

	if err := mr.db.Conn.
		Where("world_id = ? OR world_id = ?", worldID, uuid.Nil).
		Order("materials.world_id ASC, materials.id ASC").
		Offset(offset).
		Limit(limit).
		Find(&materials).Error; err != nil {
		return nil, err
	}

	return materials, nil
}

func (mr *materialsRepository) DeleteMaterial(material *models.Material) error {
	if err := mr.db.Conn.Delete(&models.Material{}, material.ID).Error; err != nil {
		return err
	}
	return nil
}
