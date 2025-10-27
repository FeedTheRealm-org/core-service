package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	SessionTokenSecretKey string
	SessionTokenDuration  time.Duration
	Dbc                   *DatabaseConfig
}

func CreateConfig() *Config {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		port = 5432
	}

	SessionTokenDuration, err := time.ParseDuration(os.Getenv("SESSION_TOKEN_DURATION"))
	if err != nil {
		SessionTokenDuration = time.Hour * 24
	}

	dbc := NewDatabaseConfig(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		port,
		os.Getenv("DB_NAME"),
	)

	return &Config{
		SessionTokenSecretKey: os.Getenv("SESSION_TOKEN_SECRET_KEY"),
		SessionTokenDuration:  SessionTokenDuration,
		Dbc:                   dbc,
	}
}
