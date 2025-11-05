package seed_database

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"golang.org/x/crypto/bcrypt"
)

func seedAccounts(db *config.DB) error {
	accounts := []models.User{
		{
			Email:    "test1@email.com",
			Password: "Password123",
			Verified: true,
		},
		{
			Email:    "test2@email.com",
			Password: "Password123",
			Verified: true,
		},
		{
			Email:    "test3@email.com",
			Password: "Password123",
			Verified: true,
		},
	}

	for _, account := range accounts {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		account.Password = string(hashedPassword)

		if err := db.Conn.Create(account).Error; err != nil {
			return err
		}
	}

	return nil
}
