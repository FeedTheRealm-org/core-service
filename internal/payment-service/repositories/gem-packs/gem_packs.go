package gem_packs

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type gemPacksRepository struct {
	conf *config.Config
	db   *config.DB
}

func NewGemPacksRepository(conf *config.Config, db *config.DB) GemPacksRepository {
	return &gemPacksRepository{
		conf: conf,
		db:   db,
	}
}

func (pr *gemPacksRepository) CreateGemPack(pack *models.GemPack) (*models.GemPack, error) {
	if err := pr.db.Conn.Create(pack).Error; err != nil {
		return nil, err
	}
	return pack, nil
}

func (pr *gemPacksRepository) GetAllGemPacks() ([]*models.GemPack, error) {
	var pack []*models.GemPack
	if err := pr.db.Conn.Find(&pack).Error; err != nil {
		return nil, err
	}
	return pack, nil
}

func (pr *gemPacksRepository) GetGemPackById(id uuid.UUID) (*models.GemPack, error) {
	var pack models.GemPack
	if err := pr.db.Conn.Where("id = ?", id).First(&pack).Error; err != nil {
		if errors.IsRecordNotFound(err) {
			return nil, errors.NewNotFoundError("package not found")
		}
		return nil, err
	}
	return &pack, nil
}

func (pr *gemPacksRepository) UpdateGemPack(id uuid.UUID, updatedPack *models.GemPack) error {
	return pr.db.Conn.Model(&models.GemPack{}).Where("id = ?", id).Updates(updatedPack).Error
}

func (pr *gemPacksRepository) DeleteGemPack(id uuid.UUID) error {
	return pr.db.Conn.Where("id = ?", id).Delete(&models.GemPack{}).Error
}
