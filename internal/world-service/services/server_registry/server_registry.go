package server_registry

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/google/uuid"
	consul_api "github.com/hashicorp/consul/api"
	nomad_api "github.com/hashicorp/nomad/api"
)

type serverRegistryService struct {
	conf         *config.Config
	nomadClient  *nomad_api.Client
	consulClient *consul_api.Client
}

func NewServerRegistryService(conf *config.Config) (ServerRegistryService, error) {
	nomadConfig := nomad_api.DefaultConfig()
	nomadConfig.Address = conf.NomadAddr
	nomadConfig.SecretID = conf.NomadToken
	nomadConfig.TLSConfig = &nomad_api.TLSConfig{
		CACert: conf.NomadCertPath,
	}
	nomadClient, err := nomad_api.NewClient(nomadConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create nomad client: %w", err)
	}

	consulConfig := consul_api.DefaultConfig()
	consulConfig.Address = conf.ConsulAddr
	consulClient, err := consul_api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &serverRegistryService{
		conf:         conf,
		nomadClient:  nomadClient,
		consulClient: consulClient,
	}, nil
}

func (s *serverRegistryService) StartNewJob(worldId uuid.UUID, zoneId int, isTest bool) error {
	templateBytes, err := os.ReadFile(s.conf.NomadTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read nomad template file: %w", err)
	}

	jobTemplate, err := template.New("ftr-server-job").Parse(string(templateBytes))
	if err != nil {
		return fmt.Errorf("failed to parse nomad template: %w", err)
	}

	jobName := fmt.Sprintf("zone-server-%s-%d", worldId.String(), zoneId)
	templateData := struct {
		JobName     string
		WorldID     string
		ZoneID      int
		IsTestWorld string
		ImageName   string
		DeployedAt  string
	}{
		JobName:     jobName,
		WorldID:     worldId.String(),
		ZoneID:      zoneId,
		IsTestWorld: strconv.FormatBool(isTest),
		ImageName:   s.conf.FTRServerImage,
		DeployedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	var rendered bytes.Buffer
	if err := jobTemplate.Execute(&rendered, templateData); err != nil {
		return fmt.Errorf("failed to render nomad template: %w", err)
	}

	job, err := s.nomadClient.Jobs().ParseHCL(rendered.String(), true)
	if err != nil {
		return fmt.Errorf("failed to parse rendered nomad job: %w", err)
	}

	_, _, err = s.nomadClient.Jobs().Register(job, nil)
	if err != nil {
		return fmt.Errorf("failed to register nomad job %q: %w", jobName, err)
	}

	return nil
}

func (s *serverRegistryService) StopJob(worldId uuid.UUID, zoneId int) error {
	jobName := fmt.Sprintf("zone-server-%s-%d", worldId.String(), zoneId)

	_, _, err := s.nomadClient.Jobs().Deregister(jobName, true, nil)
	if err != nil {
		return fmt.Errorf("failed to deregister nomad job %q: %w", jobName, err)
	}

	return nil
}

func (s *serverRegistryService) GetServerAddress(worldId uuid.UUID, zoneId int) (string, int, error) {
	filter := fmt.Sprintf(
		`"world-%s" in Service.Tags and "zone-%d" in Service.Tags`,
		worldId.String(),
		zoneId,
	)

	services, _, err := s.consulClient.Health().Service("zone-server", "", true, &consul_api.QueryOptions{
		Filter: filter,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to query consul: %w", err)
	}

	if len(services) == 0 {
		return "", 0, fmt.Errorf("no healthy server found for world %s zone %d", worldId, zoneId)
	}

	svc := services[0]
	publicIP, ok := svc.Service.Meta["public_ip"]
	if !ok || publicIP == "" {
		return "", 0, fmt.Errorf("server found but missing public_ip metadata for world %s zone %d", worldId, zoneId)
	}

	return publicIP, svc.Service.Port, nil
}
