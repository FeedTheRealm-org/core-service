package services_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
)

var service services.AccountService

func CreateStartAccountService() {
	conf := config.CreateConfig()
	logger.InitLogger(false)
	db, _ := config.NewDB(conf)
	jwtManager := session.NewJWTManager(conf.SessionTokenSecretKey, conf.SessionTokenDuration)
	repo, err := repositories.NewAccountRepository(conf, db)
	if err != nil {
		panic(err)
	}

	service = services.NewAccountService(conf, repo, jwtManager)
}

func TestMain(m *testing.M) {
	CreateStartAccountService()
	os.Exit(m.Run())
}
