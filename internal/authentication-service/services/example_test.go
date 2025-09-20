package services_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/stretchr/testify/assert"
)

var service services.ExampleService

func TestMain(m *testing.M) {
	repo := repositories.NewExampleRepository(nil)
	service = services.NewExampleService(nil, repo)
	os.Exit(m.Run())
}

func TestExample(t *testing.T) {
	data := service.GetExampleData()
	assert.Equal(t, "IM AUTH", data)
}
