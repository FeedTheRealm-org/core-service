package world_access

import (
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/google/uuid"
)

// WorldAccessService defines world join token business operations.
type WorldAccessService interface {
	IssueWorldJoinToken(userId uuid.UUID, worldId string) (*models.WorldJoinToken, error)
	ConsumeWorldJoinToken(tokenId string) (*models.WorldJoinToken, error)
}
