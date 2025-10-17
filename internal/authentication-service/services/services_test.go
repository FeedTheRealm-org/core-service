package services_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
)

var exampleService services.ExampleService
var service services.AccountService

func CreateStartAccountService() {
	conf := config.CreateConfig()
	repo, err := repositories.NewAccountRepository(conf)
	if err != nil {
		panic(err)
	}

	service = services.NewAccountService(conf, repo)
}

func CreateStartExampleService() {
	conf := config.CreateConfig()
	repo, err := repositories.NewExampleRepository(conf)
	if err != nil {
		panic(err)
	}

	exampleService = services.NewExampleService(conf, repo)
}

func TestMain(m *testing.M) {
	CreateStartAccountService()
	CreateStartExampleService()
	os.Exit(m.Run())
}
