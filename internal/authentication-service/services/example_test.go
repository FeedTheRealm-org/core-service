package services_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/stretchr/testify/assert"
)

var service services.ExampleService

func TestMain(m *testing.M) {
	conf := config.CreateConfig()
	repo, err := repositories.NewExampleRepository(conf)
	if err != nil {
		panic(err)
	}

	service = services.NewExampleService(conf, repo)
	os.Exit(m.Run())
}

func TestExample(t *testing.T) {
	data := service.GetExampleData()
	assert.Equal(t, "IM AUTH", data)
}

func TestSumQuery(t *testing.T) {
	data := service.GetSumQuery()
	assert.Equal(t, "The sum is 2", data)
}
