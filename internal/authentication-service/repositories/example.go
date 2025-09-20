package repositories

import "github.com/FeedTheRealm-org/core-service/config"

type exampleRepository struct {
	conf *config.Config
}

func NewExampleRepository(conf *config.Config) ExampleRepository {
	return &exampleRepository{
		conf: conf,
	}
}

func (er *exampleRepository) GetExampleRecord() string {
	return "IM AUTH"
}
