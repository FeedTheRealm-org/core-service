package character

// CharacterRepository defines the interface for character-related database operations.
type CharacterRepository interface {
	// UpdateCharacterInfo handles the updating of character information.
	UpdateCharacterInfo()

	// GetCharacterInfo retrieves character information.
	GetCharacterInfo()
}
