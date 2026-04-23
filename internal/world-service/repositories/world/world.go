package world

import (
	"errors"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	world_errors "github.com/FeedTheRealm-org/core-service/internal/world-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type worldRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewWorldRepository creates a new instance of WorldRepository.
func NewWorldRepository(conf *config.Config, db *config.DB) WorldRepository {
	return &worldRepository{
		conf: conf,
		db:   db,
	}
}

// StoreWorldData stores new world data in the database.
func (r *worldRepository) StoreWorldData(newWorldData *models.WorldData) (*models.WorldData, error) {
	err := r.db.Conn.Create(newWorldData).Error
	return newWorldData, err
}

// GetWorldData retrieves information for a specific world by ID.
func (r *worldRepository) GetWorldData(worldID uuid.UUID) (*models.WorldData, error) {
	var wd models.WorldData
	err := r.db.Conn.Where("id = ?", worldID).First(&wd).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, world_errors.NewWorldNotFound(err.Error())
		}
		return nil, err
	}
	return &wd, nil
}

// UpdateWorldData updates the Data and Description of an existing world and refreshes UpdatedAt, only if owned by userId.
func (r *worldRepository) UpdateWorldData(worldID uuid.UUID, userId uuid.UUID, data []byte, description string) (*models.WorldData, error) {
	var wd models.WorldData
	if err := r.db.Conn.First(&wd, "id = ?", worldID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, world_errors.NewWorldNotFound(err.Error())
		}
		return nil, err
	}
	if wd.UserId != userId {
		return nil, errors.New("forbidden: user does not own this world")
	}
	wd.Data = datatypes.JSON(data)
	wd.Description = description
	wd.UpdatedAt = time.Now().UTC()
	if err := r.db.Conn.Save(&wd).Error; err != nil {
		return nil, err
	}
	return &wd, nil
}

func (r *worldRepository) UpdateCreateableData(worldID uuid.UUID, userId uuid.UUID, createableData []byte) (*models.WorldData, error) {
	var wd models.WorldData
	if err := r.db.Conn.First(&wd, "id = ?", worldID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, world_errors.NewWorldNotFound(err.Error())
		}
		return nil, err
	}
	if wd.UserId != userId {
		return nil, errors.New("forbidden: user does not own this world")
	}
	wd.CreateableData = datatypes.JSON(createableData)
	wd.UpdatedAt = time.Now().UTC()
	if err := r.db.Conn.Save(&wd).Error; err != nil {
		return nil, err
	}
	return &wd, nil
}

func (r *worldRepository) UpsertWorldZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error) {
	var wz models.WorldZone
	err := r.db.Conn.Where("world_id = ? AND id = ?", worldID, zoneID).First(&wz).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		wz = models.WorldZone{
			ID:       zoneID,
			WorldID:  worldID,
			ZoneData: datatypes.JSON(zoneData),
		}
		if err := r.db.Conn.Create(&wz).Error; err != nil {
			return nil, err
		}
		return &wz, nil
	}

	wz.ZoneData = datatypes.JSON(zoneData)
	if err := r.db.Conn.Save(&wz).Error; err != nil {
		return nil, err
	}

	return &wz, nil
}

func (r *worldRepository) DeleteWorldData(worldID uuid.UUID) error {
	result := r.db.Conn.Delete(&models.WorldData{}, "id = ?", worldID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return world_errors.NewWorldNotFound("world not found with ID: " + worldID.String())
	}
	return nil
}

// GetWorldsList retrieves a paginated list of worlds.
func (r *worldRepository) GetWorldsList(offset int, limit int, filter string) ([]*models.WorldData, error) {
	var worlds []*models.WorldData
	query := r.db.Conn.Offset(offset).Limit(limit)

	if filter != "" {
		query = query.Where("name ILIKE ?", "%"+filter+"%")
	}

	err := query.Find(&worlds).Error
	return worlds, err
}

func (r *worldRepository) GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error) {
	var worldZones []*models.WorldZone
	if err := r.db.Conn.Where("world_id = ?", worldID).Order("id ASC").Find(&worldZones).Error; err != nil {
		return nil, err
	}

	return worldZones, nil
}

func (r *worldRepository) GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error) {
	var worldZone models.WorldZone
	if err := r.db.Conn.Where("world_id = ? AND id = ?", worldID, zoneID).First(&worldZone).Error; err != nil {
		return nil, err
	}
	return &worldZone, nil
}

func (r *worldRepository) ClearDatabase() error {
	if err := r.db.Conn.Delete(&models.WorldZone{}, "1 = 1").Error; err != nil {
		return err
	}
	err := r.db.Conn.Delete(&models.WorldData{}, "1 = 1").Error
	return err
}

func (r *worldRepository) GetUserIdByWorldId(worldID uuid.UUID) (uuid.UUID, error) {
	var world models.WorldData
	err := r.db.Conn.Select("user_id").Where("id = ?", worldID).First(&world).Error
	return world.UserId, err
}

func (r *worldRepository) GetTotalZonesCountByUserId(userId uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Conn.Table("world_zones wz").
		Joins("INNER JOIN world_data wd ON wz.world_id = wd.id").
		Where("wd.user_id = ?", userId).
		Count(&count).Error
	return count, err
}
