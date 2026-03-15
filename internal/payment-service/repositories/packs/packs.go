package packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type packsRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewPacksRepository(conf *config.Config, db *config.DB) PacksRepository {
	return &packsRepository{
		conf: conf,
		db:   db,
	}
}

func (pr *packsRepository) CreatePack(pack *models.Pack) (*models.Pack, error) {
	if err := pr.db.Conn.Create(pack).Error; err != nil {
		return nil, err
	}
	return pack, nil
}

func (pr *packsRepository) GetAllPacks() ([]*models.Pack, error) {
	var pack []*models.Pack
	if err := pr.db.Conn.Find(&pack).Error; err != nil {
		return nil, err
	}
	return pack, nil
}

func (pr *packsRepository) GetPackById(id uuid.UUID) (*models.Pack, error) {
	var pack models.Pack
	if err := pr.db.Conn.Where("id = ?", id).First(&pack).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, errors.NewNotFoundError("package not found")
		}
		return nil, err
	}
	return &pack, nil
}

func (pr *packsRepository) UpdatePack(id uuid.UUID, updatedPack *models.Pack) (*models.Pack, error) {
	if err := pr.db.Conn.Model(&models.Pack{}).Where("id = ?", id).Updates(updatedPack).Error; err != nil {
		return nil, err
	}
	return updatedPack, nil
}

func (pr *packsRepository) DeletePack(id uuid.UUID) error {
	return pr.db.Conn.Where("id = ?", id).Delete(&models.Pack{}).Error
}
