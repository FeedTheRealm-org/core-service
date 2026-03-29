package gem_packs

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type GemPacksRepository interface {
	CreateGemPack(pkg *models.GemPack) (*models.GemPack, error)
	GetAllGemPacks() ([]*models.GemPack, error)
	GetGemPackById(id uuid.UUID) (*models.GemPack, error)
	UpdateGemPack(id uuid.UUID, updatedPkg *models.GemPack) error
	DeleteGemPack(id uuid.UUID) error
}
