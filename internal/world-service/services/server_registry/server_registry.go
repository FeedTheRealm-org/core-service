package server_registry

import (
	"github.com/google/uuid"
)

type serverRegistryService struct{}

func NewServerRegistryService() ServerRegistryService {
	return &serverRegistryService{}
}

func (sr serverRegistryService) RegisterServer(worldId uuid.UUID, zoneId int, address string) {}

func (sr serverRegistryService) UnRegisterServer(worldId uuid.UUID, zoneId int) {}
