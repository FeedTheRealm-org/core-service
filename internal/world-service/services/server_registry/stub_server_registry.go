package server_registry

import (
	"github.com/google/uuid"
)

type stubServerRegistryService struct {
}

// NewWorldService creates a new instance of WorldService.
func NewStubServerRegistryService() ServerRegistryService {
	return &stubServerRegistryService{}
}

func (ns *stubServerRegistryService) StartNewJob(worldId uuid.UUID, zoneId int, isTest bool) error {
	return nil
}

func (ns *stubServerRegistryService) StopJob(worldId uuid.UUID, zoneId int) error {
	return nil
}

// When testing, make sure to match the port with the zone via this relation: port = 7776 + zoneId
func (ns *stubServerRegistryService) GetServerAddress(worldId uuid.UUID, zoneId int) (string, int, error) {
	return "127.0.0.1", 7776 + zoneId, nil
}
