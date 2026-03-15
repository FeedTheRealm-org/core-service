package packs

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type PacksRepository interface {
	CreatePack(pkg *models.Pack) (*models.Pack, error)
	GetAllPacks() ([]*models.Pack, error)
	GetPackById(id uuid.UUID) (*models.Pack, error)
	UpdatePack(id uuid.UUID, updatedPkg *models.Pack) (*models.Pack, error)
	DeletePack(id uuid.UUID) error
}
