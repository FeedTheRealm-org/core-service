package server_registry

import (
	"github.com/google/uuid"
)

// ServerRegistryService defines the interface for nomad job api calls
type ServerRegistryService interface {
	// Starts new or restarts nomad service for a world and zone, this will be called when a world is published
	StartNewJob(worldId uuid.UUID, zoneId int, isTest bool) error

	// StopJob stops the nomad job for a world and zone, this will be called when a world is unpublished or deleted
	StopJob(worldId uuid.UUID, zoneId int) error

	// GetServerAddress returns the IP and port of the server running the world - zone
	GetServerAddress(worldId uuid.UUID, zoneId int) (string, int, error)
}
