package character

import "github.com/gin-gonic/gin"

// CharacterController defines the interface for character-related HTTP operations.
type CharacterController interface {
	// PatchCharacterInfo handles the updating of character information.
	PatchCharacterInfo(c *gin.Context)

	// GetCharacterInfo retrieves character information.
	GetCharacterInfo(c *gin.Context)
}
