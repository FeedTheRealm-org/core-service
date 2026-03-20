package server_registry

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/google/uuid"
	consul_api "github.com/hashicorp/consul/api"
	nomad_api "github.com/hashicorp/nomad/api"
)

type serverRegistryService struct {
	conf *config.Config
}

// NewWorldService creates a new instance of WorldService.
func NewServerRegistryService(conf *config.Config) ServerRegistryService {
	return &serverRegistryService{
		conf: conf,
	}
}

func (s *serverRegistryService) StartNewJob(worldId uuid.UUID, zoneId int) error {
	if s.conf.NomadAddr == "" {
		return fmt.Errorf("nomad address is empty")
	}

	templateBytes, err := os.ReadFile(s.conf.NomadTemplatePath)
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
		ImageName: s.conf.FTRServerImage,
	}

	var rendered bytes.Buffer
	if err := jobTemplate.Execute(&rendered, templateData); err != nil {
		return fmt.Errorf("failed to render nomad template: %w", err)
	}

	apiConfig := nomad_api.DefaultConfig()
	apiConfig.Address = s.conf.NomadAddr
	apiConfig.SecretID = s.conf.NomadToken
	apiConfig.TLSConfig = &nomad_api.TLSConfig{
		CACert: s.conf.NomadCertPath,
	}

	client, err := nomad_api.NewClient(apiConfig)
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

func (s *serverRegistryService) GetServerAddress(worldId uuid.UUID, zoneId int) (string, int, error) {
	consulConfig := consul_api.DefaultConfig()

	client, err := consul_api.NewClient(consulConfig)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create consul client: %w", err)
	}

	filter := fmt.Sprintf(
		`"world-%s" in Service.Tags and "zone-%d" in Service.Tags`,
		worldId.String(),
		zoneId,
	)

	services, _, err := client.Health().Service("zone-server", "", true, &consul_api.QueryOptions{
		Filter: filter,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to query consul: %w", err)
	}

	if len(services) == 0 {
		return "", 0, fmt.Errorf("no healthy server found for world %s zone %d", worldId, zoneId)
	}

	s_ := services[0]
	publicIP, ok := s_.Service.Meta["public_ip"]
	if !ok || publicIP == "" {
		return "", 0, fmt.Errorf("server found but missing public_ip metadata for world %s zone %d", worldId, zoneId)
	}

	return publicIP, s_.Service.Port, nil
}
