package repositories_test

import (
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/stretchr/testify/assert"
)

func TestAccountRepository_CreateAccount(t *testing.T) {
	conf := config.CreateConfig()
	logger.InitLogger(false)
	db, _ := config.NewDB(conf)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := "john.doe@example.com"
	passwordHash := "hashed_password"

	user := &repositories.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	err = repo.CreateAccount(user)
	assert.Nil(t, err, "failed to create account")

	result, err := repo.GetAccountByEmail(email)
	assert.Nil(t, err, "failed to get account by email")
	assert.NotNil(t, result, "expected user, got nil")
	assert.Equal(t, email, result.Email, "unexpected email")
	assert.Equal(t, passwordHash, result.PasswordHash, "unexpected password hash")
}

func TestAccountRepository_GetAccountByEmail_NotFound(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := "notfound@example.com"
	user, err := repo.GetAccountByEmail(email)
	assert.NotNil(t, err, "expected error on getting non-existing user")
	assert.Error(t, err, "Account not found")
	assert.Nil(t, user, "expected no user to be found")
}
