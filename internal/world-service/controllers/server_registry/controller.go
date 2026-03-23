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
}
