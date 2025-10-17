package repositories_test

import (
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
)

func TestAccountRepository_CreateAccount(t *testing.T) {
	conf := config.CreateConfig()
	repo, err := repositories.NewAccountRepository(conf)
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	email := "john.doe@example.com"
	passwordHash := "hashed_password"

	user := &repositories.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	err = repo.CreateAccount(user)
	result, err := repo.GetAccountByEmail(email)

	if result == nil {
		t.Errorf("expected user, got nil")
	}

	if result.Email != email || result.PasswordHash != passwordHash {
		t.Errorf("expected user with email %s and password hash %s, got email %s and password hash %s",
			email,
			passwordHash,
			result.Email,
			result.PasswordHash,
		)
	}
}
