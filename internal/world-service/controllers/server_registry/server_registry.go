package server_registry

import "github.com/gin-gonic/gin"

// ServerRegistryController handles ftr-server reports and WorldId,ZoneId -> IP:port mapping
type serverRegistryController struct {
}

func NewServerRegistryController() ServerRegistryController {
	return &serverRegistryController{}
}

func (sr *serverRegistryController) RegisterServer(c *gin.Context) {

}

func (sr *serverRegistryController) UnRegisterServer(c *gin.Context) {

}
