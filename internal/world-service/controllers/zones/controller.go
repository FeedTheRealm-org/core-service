package zones

import "github.com/gin-gonic/gin"

// ZonesController defines the interface for zone-related HTTP operations.
type ZonesController interface {
	// PublishZone publishes or updates a zone for a world.
	PublishZone(c *gin.Context)

	// GetWorldZones retrieves available zones for a specific world.
	GetWorldZones(c *gin.Context)

	// GetWorldZoneData retrieves specific zone data for a specific world.
	GetWorldZoneData(c *gin.Context)

	// ActivateZone starts server orchestration for a published zone.
	ActivateZone(c *gin.Context)

	// DeactivateZone stops server orchestration for an active zone.
	DeactivateZone(c *gin.Context)
}
