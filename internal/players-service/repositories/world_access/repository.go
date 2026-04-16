package world_access

import (
	"time"

	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/google/uuid"
)

// WorldAccessRepository defines persistence operations for world join tokens.
type WorldAccessRepository interface {
	CreateWorldJoinToken(token *models.WorldJoinToken) error
	ConsumeWorldJoinToken(tokenId uuid.UUID, now time.Time) (*models.WorldJoinToken, error)
}
