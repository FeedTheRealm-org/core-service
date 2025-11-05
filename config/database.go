package config

import (
	"context"
	"fmt"

	"time"

	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	dsn  string
	Conn *pgxpool.Pool
}

func NewDB(conf *Config) (*DB, error) {
	dsn := generateURL(conf.DB)

	var conn *pgxpool.Pool
	var err error
	for range conf.DB.ConnectionRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, err = pgxpool.New(ctx, dsn)
		if err == nil {
			err = conn.Ping(ctx)
			if err == nil {
				break
			}
		}

		logger.Logger.Warnf("Failed to connect to the database: %v. Retrying in 1 second...", err)
		time.Sleep(1 * time.Second)
	}

	db := &DB{
		dsn:  dsn,
		Conn: conn,
	}

	if conf.DB.ShouldMigrate {
		err = db.runMigrations()
		if err != nil {
			return nil, err
		}
	}

	logger.Logger.Infoln("Connected to the database & migrations applied")
	return db, nil
}

func (db *DB) runMigrations() error {
	m, err := migrate.New(
		"file://./migrations",
		db.dsn,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

/* --- UTILS --- */

func generateURL(dbc *DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbc.Username,
		dbc.Password,
		dbc.Host,
		dbc.Port,
		dbc.Database,
	)
}
