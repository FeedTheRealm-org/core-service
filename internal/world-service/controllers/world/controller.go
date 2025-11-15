package world

import "github.com/gin-gonic/gin"

// WorldController defines the interface for world-related HTTP operations.
type WorldController interface {
	// PublishWorld handles the publishing of world information.
	PublishWorld(c *gin.Context)

	// GetWorldData retrieves information for a specific world.
	GetWorld(c *gin.Context)

	// GetWorldsList retrieves a paginated list of worlds.
	GetWorldsList(c *gin.Context)

	// ResetDatabase clears all data in the database.
	ResetDatabase(c *gin.Context)
}
