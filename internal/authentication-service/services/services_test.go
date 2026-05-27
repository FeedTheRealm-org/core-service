package services_test

import (
	"os"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/utils/email_sender"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
)

var accountService services.AccountService
var emailSenderService email_sender.EmailSenderService

func CreateStartAccountService() {
	conf := config.CreateConfig()
	logger.InitLogger(false)
	db, _ := config.NewDB(conf)
	jwtManager := session.NewJWTManager(conf.SessionAccessTokenSecretKey, conf.SessionRefreshTokenSecretKey, conf.SessionAccessTokenDuration, conf.SessionRefreshTokenDuration)
	repo, err := repositories.NewAccountRepository(conf, db)
	if err != nil {
		panic(err)
	}

	accountService = services.NewAccountService(conf, repo, jwtManager)
}

func CreateStartEmailSenderService() {
	conf := config.CreateConfig()
	emailSenderService = email_sender.NewEmailSenderService(conf)
}

func clearDatabase(db *config.DB) {
	_ = db.Conn.Exec("DELETE FROM account_verifications WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');")
	_ = db.Conn.Exec("DELETE FROM users WHERE email LIKE '%@example.com';")
}

func TestMain(m *testing.M) {
	if err := os.Setenv("SERVER_FIXED_TOKEN", "test-fixed-token"); err != nil {
		panic(err)
	}
	CreateStartAccountService()
	CreateStartEmailSenderService()

	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	clearDatabase(db)

	code := m.Run()

	clearDatabase(db)
	os.Exit(code)
}
