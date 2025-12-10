package bundles

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// BundlesRepository defines the interface for models-related database operations.
type BundlesRepository interface {
	PublishWorldBundle(models models.Bundle) (models.Bundle, error)
	GetWorldBundle(worldId uuid.UUID) (models.Bundle, error)
}
