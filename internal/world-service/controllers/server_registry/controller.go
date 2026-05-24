package server_registry

import "github.com/gin-gonic/gin"

// ServerRegistryController handles ftr-server reports and WorldId,ZoneId -> IP:port mapping
type ServerRegistryController interface {
	// StartNewJob starts a new server by id as ADMIN.
	StartNewJob(c *gin.Context)

	// StartNewJob starts a new server by id as ADMIN.
	StopJob(c *gin.Context)

	// UnRegisterServer removes the server entry.
	GetServerAddress(c *gin.Context)

	// UpdateServer is a webhook endpoint for reload server when image is updated.
	UpdateServer(c *gin.Context)

	// UpdateStatus is a webhook endpoint for updating server status.
	UpdateStatus(c *gin.Context)

	// UpdatePlayerCount is a webhook endpoint for player count updates.
	UpdatePlayerCount(c *gin.Context)

	// GetWorldPlayerCounts returns the current player counts for a world.
	GetWorldPlayerCounts(c *gin.Context)

	// GetAllWorldPlayerCounts returns the player counts for all worlds.
	GetAllWorldPlayerCounts(c *gin.Context)
}
