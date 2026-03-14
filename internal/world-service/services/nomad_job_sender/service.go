package nomad_job_sender

import (
	"github.com/google/uuid"
)

// NomadJobSenderService defines the interface for nomad job api calls
type NomadJobSenderService interface {
	// Starts new nomad service for a world and zone, this will be called when a world is published
	StartNewJob(worldId uuid.UUID, zoneId int) error
}
