package services_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
)

var account_service services.AccountService
var email_sender_service services.EmailSenderService

func CreateStartAccountService() {
	conf := config.CreateConfig()
	logger.InitLogger(false)
	db, _ := config.NewDB(conf)
	repo, err := repositories.NewAccountRepository(conf, db)
	if err != nil {
		panic(err)
	}

	account_service = services.NewAccountService(conf, repo)
}

func CreateStartEmailSenderService() {
	conf := config.CreateConfig()
	email_sender_service = services.NewEmailSenderService(conf)
}

func TestMain(m *testing.M) {
	CreateStartAccountService()
	CreateStartEmailSenderService()
	os.Exit(m.Run())
}
