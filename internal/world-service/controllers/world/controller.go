package world

import "github.com/gin-gonic/gin"

// WorldController defines the interface for world-related HTTP operations.
type WorldController interface {
	// PublishWorld handles the publishing of world information.
	PublishWorld(c *gin.Context)

	// PublishZone publishes or updates a zone for a world.
	PublishZone(c *gin.Context)

	// GetWorldData retrieves information for a specific world.
	GetWorld(c *gin.Context)

	// GetWorldZones retrieves available zones for a specific world.
	GetWorldZones(c *gin.Context)

	// GetWorldsList retrieves a paginated list of worlds.
	GetWorldsList(c *gin.Context)

	// UpdateWorld updates the data and description for an existing world.
	UpdateWorld(c *gin.Context)

	// UpdateCreateableData updates createable data for an existing world.
	UpdateCreateableData(c *gin.Context)

	// DeleteWorld handles the deletion of a world.
	DeleteWorld(c *gin.Context)

	// ResetDatabase clears all data in the database.
	ResetDatabase(c *gin.Context)
}
