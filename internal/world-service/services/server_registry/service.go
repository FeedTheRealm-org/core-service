package server_registry

import (
	"github.com/google/uuid"
)

// ServerRegistryService defines the interface for nomad job api calls
type ServerRegistryService interface {
	// Starts new nomad service for a world and zone, this will be called when a world is published
	StartNewJob(worldId uuid.UUID, zoneId int) error

	GetServerAddress(worldId uuid.UUID, zoneId int) (string, int, error)
}
