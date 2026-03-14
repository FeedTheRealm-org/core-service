package nomad_job_sender

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/google/uuid"
	"github.com/hashicorp/nomad/api"
)

type nomadJobSenderService struct {
	conf *config.Config
}

// NewWorldService creates a new instance of WorldService.
func NewNomadJobSenderService(conf *config.Config) NomadJobSenderService {
	return &nomadJobSenderService{
		conf: conf,
	}
}

func (ns *nomadJobSenderService) StartNewJob(worldId uuid.UUID, zoneId int) error {
	if ns.conf.NomadAddr == "" {
		return fmt.Errorf("nomad address is empty")
	}

	templateBytes, err := os.ReadFile(ns.conf.NomadTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read nomad template file: %w", err)
	}

	// TODO: add resources cap
	jobTemplate, err := template.New("ftr-server-job").Parse(string(templateBytes))
	if err != nil {
		return fmt.Errorf("failed to parse nomad template: %w", err)
	}

	jobName := fmt.Sprintf("zone-server-%s-%d", worldId.String(), zoneId)
	templateData := struct {
		JobName   string
		WorldID   string
		ZoneID    int
		ImageName string
	}{
		JobName:   jobName,
		WorldID:   worldId.String(),
		ZoneID:    zoneId,
		ImageName: ns.conf.FTRServerImage,
	}

	var rendered bytes.Buffer
	if err := jobTemplate.Execute(&rendered, templateData); err != nil {
		return fmt.Errorf("failed to render nomad template: %w", err)
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = ns.conf.NomadAddr

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return fmt.Errorf("failed to create nomad client: %w", err)
	}

	job, err := client.Jobs().ParseHCL(rendered.String(), true)
	if err != nil {
		return fmt.Errorf("failed to parse rendered nomad job: %w", err)
	}

	_, _, err = client.Jobs().Register(job, nil)
	if err != nil {
		return fmt.Errorf("failed to register nomad job %q: %w", jobName, err)
	}

	return nil
}
