package main

import (
	"fmt"
	"os"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.CreateConfig()
	log := logger.InitLogger(true)
	dbURL := generateURL(cfg.DB)

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatal(err)
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migrations rolled back successfully")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
	default:
		log.Fatalf("Unknown command: %s. Use 'up', 'down', or 'version'", command)
	}
}

func generateURL(dbc *config.DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbc.Username,
		dbc.Password,
		dbc.Host,
		dbc.Port,
		dbc.Database,
	)
}
