package server_registry

import (
	"github.com/google/uuid"
)

type serverRegistryRepository struct{}

func NewServerRegistryRepository() ServerRegistryRepository {
	return &serverRegistryRepository{}
}

func (sr serverRegistryRepository) RegisterServer(worldId uuid.UUID, zoneId int, address string) {}

func (sr serverRegistryRepository) UnRegisterServer(worldId uuid.UUID, zoneId int) {}
