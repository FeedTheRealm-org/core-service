package config

import (
	"os"
	"strconv"
)

type Config struct {
	Dbc *DatabaseConfig
}

func CreateConfig() *Config {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		port = 5432
	}

	dbc := NewDatabaseConfig(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		port,
		os.Getenv("DB_NAME"),
	)

	return &Config{
		Dbc: dbc,
	}
}
