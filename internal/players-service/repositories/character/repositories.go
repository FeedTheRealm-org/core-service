package character

import (
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/google/uuid"
)

// CharacterRepository defines the interface for character-related database operations.
type CharacterRepository interface {
	// UpdateCharacterInfo handles the updating of character information.
	UpdateCharacterInfo(newCharacterInfo *models.CharacterInfo) error

	// GetCharacterInfo retrieves character information.
	GetCharacterInfo(userId uuid.UUID) (*models.CharacterInfo, error)
}
