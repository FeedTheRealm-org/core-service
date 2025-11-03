package seed_database

import (
	"github.com/FeedTheRealm-org/core-service/config"
)

func SeedDatabase(db *config.DB) error {
	err := seedAccounts(db)
	if err != nil {
		return err
	}
	return nil
}
