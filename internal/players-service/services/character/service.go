package character

import (
	"github.com/FeedTheRealm-org/core-service/internal/players-service/models"
	"github.com/google/uuid"
)

// CharacterService defines the interface for character-related operations.
type CharacterService interface {
	// UpdateCharacterInfo handles the updating of character information.
	UpdateCharacterInfo(userId uuid.UUID, newCharacterInfo *models.CharacterInfo, newCategorySprites []models.CategorySprite) error

	// GetCharacterInfo retrieves character information.
	GetCharacterInfo(userId uuid.UUID) (*models.CharacterInfo, []models.CategorySprite, error)
}
