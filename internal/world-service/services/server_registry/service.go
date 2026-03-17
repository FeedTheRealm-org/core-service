package server_registry

import (
	"github.com/google/uuid"
)

// ServerRegistryService business logic for server registration.
type ServerRegistryService interface {
	// RegisterServer registers server in world-service.
	RegisterServer(worldId uuid.UUID, zoneId int, address string)

	// UnRegisterServer removes the server entry.
	UnRegisterServerByWorld(worldId uuid.UUID, zoneId int)

	// UnRegisterServer removes the server entry.
	UnRegisterServerById(serverId uuid.UUID)
}
