package packs

import (
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/google/uuid"
)

type PacksService interface {
	// GetAllPacks retrieves a list of available packs.
	GetAllPacks() ([]*models.Pack, error)

	// GetPackById retrieves a pack by its ID.
	GetPackById(packageId uuid.UUID) (*models.Pack, error)

	// CreatePack creates a new pack with the provided details.
	CreatePack(name string, gems int, price float32) (*models.Pack, error)

	// UpdatePack updates the details of an existing pack.
	UpdatePack(packageId uuid.UUID, name string, gems int, price float32) (*models.Pack, error)

	// DeletePack deletes a pack by its ID.
	DeletePack(packageId uuid.UUID) error
}
