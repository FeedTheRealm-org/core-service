package config

import (
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Port            int
	ShutdownTimeout time.Duration
}

type Config struct {
	SessionTokenSecretKey string
	SessionTokenDuration  time.Duration
	Server                *ServerConfig
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

	ShutdownTimeout, err := time.ParseDuration(os.Getenv("SERVER_SHUTDOWN_TIMEOUT"))
	if err != nil {
		ShutdownTimeout = time.Second * 30
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
		Server: &ServerConfig{
			Port:            8000,
			ShutdownTimeout: ShutdownTimeout,
		},
		Dbc: dbc,
	}
}
