package seed_database

import (
	"context"

	"github.com/FeedTheRealm-org/core-service/config"
	"golang.org/x/crypto/bcrypt"
)

func seedAccounts(db *config.DB) error {
	accounts := []struct {
		Email    string
		Password string
	}{
		{
			Email:    "test1@email.com",
			Password: "Password123",
		},
		{
			Email:    "test2@email.com",
			Password: "Password123",
		},
		{
			Email:    "test3@email.com",
			Password: "Password123",
		},
	}

	for _, account := range accounts {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = db.Conn.Exec(context.Background(),
			`INSERT INTO accounts (email, password_hash)
			VALUES ($1, $2)`, account.Email, hashedPassword)
		if err != nil {
			return err
		}
	}

	return nil
}
