package config

import (
	"os"
	"strconv"
	"time"
)

type EnvironmentType int

const (
	Development EnvironmentType = iota
	Testing
	Production
)

type ServerConfig struct {
	Port            int
	ShutdownTimeout time.Duration
	Environment     EnvironmentType
}

type DatabaseConfig struct {
	Username          string
	Password          string
	Host              string
	Port              int
	Database          string
	ConnectionRetries int
	ShouldMigrate     bool
}

type Config struct {
	Server                *ServerConfig
	DB                    *DatabaseConfig
	SessionTokenSecretKey string
	SessionTokenDuration  time.Duration
	BrevoAPIKey           string
	EmailSenderAddress    string
}

func CreateConfig() *Config {
	dbc := &DatabaseConfig{
		Username:          getEnvOrDefaultString("DB_USERNAME", "postgres"),
		Password:          getEnvOrDefaultString("DB_PASSWORD", "postgres"),
		Host:              getEnvOrDefaultString("DB_HOST", "localhost"),
		Port:              getEnvOrDefaultInt("DB_PORT", 5432),
		Database:          getEnvOrDefaultString("DB_NAME", "core_service"),
		ConnectionRetries: getEnvOrDefaultInt("DB_CONNECTION_RETRIES", 10),
		ShouldMigrate:     getEnvOrDefaultString("DB_SHOULD_MIGRATE", "false") == "true",
	}

	return &Config{
		Server: &ServerConfig{
			Port:            getEnvOrDefaultInt("SERVER_PORT", 8000),
			ShutdownTimeout: getEnvOrDefaultDuration("SERVER_SHUTDOWN_TIMEOUT", time.Second*30),
			Environment:     getEnvironmentType(os.Getenv("SERVER_ENVIRONMENT")),
		},
		DB:                    dbc,
		SessionTokenSecretKey: os.Getenv("SESSION_TOKEN_SECRET_KEY"),
		SessionTokenDuration:  getEnvOrDefaultDuration("SESSION_TOKEN_DURATION", time.Hour*24),
		BrevoAPIKey:           os.Getenv("BREVO_API_KEY"),
		EmailSenderAddress:    os.Getenv("EMAIL_SENDER_ADDRESS"),
	}
}

/* --- ENV Getters --- */

func getEnvOrDefaultString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvironmentType(env string) EnvironmentType {
	switch env {
	case "development":
		return Development
	case "testing":
		return Testing
	case "production":
		return Production
	default:
		return Development
	}
}
