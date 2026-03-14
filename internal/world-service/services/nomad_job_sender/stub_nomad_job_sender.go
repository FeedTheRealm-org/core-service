package nomad_job_sender

import (
	"github.com/google/uuid"
)

type stubNomadJobSenderService struct{}

// NewWorldService creates a new instance of WorldService.
func NewStubNomadJobSenderService() NomadJobSenderService {
	return &stubNomadJobSenderService{}
}

func (ns *stubNomadJobSenderService) StartNewJob(worldId uuid.UUID, zoneId int) error {
	return nil
}
