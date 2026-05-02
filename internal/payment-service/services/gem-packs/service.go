package gem_packs

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GemPacksService interface {
	// GetAllPacks retrieves a list of available packs.
	GetAllGemPacks() ([]*models.GemPack, error)

	// GetPackById retrieves a pack by its ID.
	GetGemPackById(packageId uuid.UUID) (*models.GemPack, error)

	// CreatePack creates a new pack with the provided details.
	CreateGemPack(name string, gems int64, price decimal.Decimal) (*models.GemPack, error)

	// UpdatePack updates only provided fields of an existing pack.
	UpdateGemPack(packageId uuid.UUID, name string, gems int64, price decimal.Decimal) (*models.GemPack, error)

	// DeletePack deletes a pack by its ID.
	DeleteGemPack(packageId uuid.UUID) error
}
