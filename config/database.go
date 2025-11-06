package config

import (
	"fmt"

	"time"

	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type DB struct {
	dsn  string
	Conn *gorm.DB
}

func NewDB(conf *Config) (*DB, error) {
	dsn := generateURL(conf.DB)

	var dbLogger gormLogger.Interface
	if conf.Server.Environment == Production {
		dbLogger = gormLogger.Default.LogMode(gormLogger.Silent)
	} else {
		dbLogger = gormLogger.Default.LogMode(gormLogger.Info)
	}

	var conn *gorm.DB
	var err error
	for range conf.DB.ConnectionRetries {
		if conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:         dbLogger,
			TranslateError: true,
		}); err == nil {
			break
		}

		logger.Logger.Warnf("Failed to connect to the database: %v. Retrying in 1 second...", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, err
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
