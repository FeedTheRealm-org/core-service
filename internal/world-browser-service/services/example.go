package services

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-browser-service/repositories"
)

type exampleService struct {
	conf *config.Config
	repo repositories.ExampleRepository
}

func NewExampleService(conf *config.Config, repo repositories.ExampleRepository) ExampleService {
	return &exampleService{
		conf: conf,
		repo: repo,
	}
}

func (es *exampleService) GetExampleData() string {
	return es.repo.GetExampleRecord()
}
