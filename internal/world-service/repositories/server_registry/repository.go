package server_registry

import (
	"github.com/google/uuid"
)

// ServerRegistryRepository business logic for server registration.
type ServerRegistryRepository interface {
	// RegisterServer registers server in world-service.
	RegisterServer(worldId uuid.UUID, zoneId int, address string)

	// UnRegisterServer removes the server entry.
	UnRegisterServer(worldId uuid.UUID, zoneId int)
}
