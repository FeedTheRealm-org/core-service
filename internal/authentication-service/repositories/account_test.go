package repositories_test

import (
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/stretchr/testify/assert"
)

func TestAccountRepository_CreateAccount(t *testing.T) {
	conf := config.CreateConfig()
	repo, err := repositories.NewAccountRepository(conf)
	assert.Nil(t, err, "failed to connect to database")

	email := "john.doe@example.com"
	passwordHash := "hashed_password"

	user := &repositories.User{
		Email:        email,
		PasswordHash: passwordHash,
		VerifyCode:   "verification_code",
		Expiration:   time.Now().Add(24 * time.Hour),
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
	repo, err := repositories.NewAccountRepository(conf)
	assert.Nil(t, err, "failed to connect to database")

	email := "notfound@example.com"
	user, err := repo.GetAccountByEmail(email)
	assert.NotNil(t, err, "expected error on getting non-existing user")
	assert.Error(t, err, "Account not found")
	assert.Nil(t, user, "expected no user to be found")
}

func TestAccountRepository_IsAccountVerified(t *testing.T) {
	conf := config.CreateConfig()
	repo, err := repositories.NewAccountRepository(conf)
	assert.Nil(t, err, "failed to connect to database")

	email := "johndoe@example.com"
	passwordHash := "hashed_password"

	user := &repositories.User{
		Email:        email,
		PasswordHash: passwordHash,
		VerifyCode:   "verification_code",
		Expiration:   time.Now().Add(24 * time.Hour),
	}

	err = repo.CreateAccount(user)
	assert.Nil(t, err, "failed to create account")

	isVerified, err := repo.IsAccountVerified(email)
	assert.Nil(t, err, "failed to check if account is verified")
	assert.False(t, isVerified, "expected account to be unverified")
}

func TestAccountRepository_VerifyAccount(t *testing.T) {
	conf := config.CreateConfig()
	repo, err := repositories.NewAccountRepository(conf)
	assert.Nil(t, err, "failed to connect to database")

	email := "johndoe@example.com"
	passwordHash := "hashed_password"
	code := "verification_code"

	user := &repositories.User{
		Email:        email,
		PasswordHash: passwordHash,
		VerifyCode:   code,
		Expiration:   time.Now().Add(24 * time.Hour),
	}

	err = repo.CreateAccount(user)
	assert.Nil(t, err, "failed to create account")

	err = repo.VerifyAccount(email, code, time.Now())
	assert.Nil(t, err, "failed to verify account")

	isVerified, err := repo.IsAccountVerified(email)
	assert.Nil(t, err, "failed to check if account is verified")
	assert.True(t, isVerified, "expected account to be verified")
}

func TestAccountRepository_VerifyAccount_Expired(t *testing.T) {
	conf := config.CreateConfig()
	repo, err := repositories.NewAccountRepository(conf)
	assert.Nil(t, err, "failed to connect to database")

	email := "johndoe_expired@example.com"
	passwordHash := "hashed_password"
	code := "verification_code"

	user := &repositories.User{
		Email:        email,
		PasswordHash: passwordHash,
		VerifyCode:   code,
		Expiration:   time.Now().Add(-time.Hour), // Already expired
	}

	err = repo.CreateAccount(user)
	assert.Nil(t, err, "failed to create account")

	err = repo.VerifyAccount(email, code, time.Now())
	assert.NotNil(t, err, "expected error on verifying expired account")
	assert.Error(t, err, "Account verification has expired")
}
