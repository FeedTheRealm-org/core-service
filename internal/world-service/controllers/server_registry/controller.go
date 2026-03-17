package server_registry

import "github.com/gin-gonic/gin"

// ServerRegistryController handles ftr-server reports and WorldId,ZoneId -> IP:port mapping
type ServerRegistryController interface {
	// RegisterServer registers server in world-service.
	RegisterServer(c *gin.Context)

	// UnRegisterServer removes the server entry.
	UnRegisterServer(c *gin.Context)
}
