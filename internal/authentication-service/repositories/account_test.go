package repositories_test

import (
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/stretchr/testify/assert"
)

func TestAccountRepository_CreateAccount(t *testing.T) {
	logger.InitLogger(false)

	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := "john.doe@example.com"
	passwordHash := "hashed_password"

	user := &models.User{
		Email:    email,
		Password: passwordHash,
	}

	err = repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")

	result, err := repo.GetAccountById(user.Id)
	assert.Nil(t, err, "failed to get account by email")
	assert.NotNil(t, result, "expected user, got nil")
	assert.Equal(t, email, result.Email, "unexpected email")
	assert.Equal(t, passwordHash, result.Password, "unexpected password hash")
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

func TestAccountRepository_IsAccountVerified(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := "johndoe@example.com"
	passwordHash := "hashed_password"

	user := &models.User{
		Email:    email,
		Password: passwordHash,
	}

	err = repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")
	assert.NotEmpty(t, user.Id.String())

	assert.Nil(t, err, "failed to check if account is verified")
	assert.False(t, user.Verified, "expected account to be unverified")
}

func TestAccountRepository_VerifyAccount(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := "johndoe@example.com"
	code := "verification_code"

	user, err := repo.GetAccountByEmail(email)
	assert.NoError(t, err, "failed to get account by email")

	err = repo.VerifyAccount(user, code, time.Now())
	assert.Nil(t, err, "failed to verify account")

	assert.Nil(t, err, "failed to check if account is verified")
	assert.True(t, user.Verified, "expected account to be verified")
}

func TestAccountRepository_VerifyAccount_Expired(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := "johndoe@example.com"
	code := "verification_code"

	user := &models.User{
		Email:    email,
		Password: "a",
	}

	err = repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")
	assert.NotEmpty(t, user.Id.String())

	err = repo.VerifyAccount(user, code, time.Now().Add(-time.Hour))
	assert.NotNil(t, err, "expected error on verifying expired account")
	assert.Error(t, err, "Account verification has expired")
}
