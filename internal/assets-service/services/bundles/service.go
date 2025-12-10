package models

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type BundleService interface {
	// PublishWorldBundle publishes a new bundle for a specific world.
	PublishWorldBundle(bundle models.Bundle) (models.Bundle, error)
	// GetWorldBundle retrieves the bundle for a specific world.
	GetWorldBundle(worldId uuid.UUID) (models.Bundle, error)
}
