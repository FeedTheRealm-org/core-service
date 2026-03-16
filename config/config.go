package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type EnvironmentType int

const (
	Development EnvironmentType = iota
	Testing
	Production
)

type ServerConfig struct {
	Hostname        string
	Port            int
	ShutdownTimeout time.Duration
	Environment     EnvironmentType
	PublicIP        string
}

type DatabaseConfig struct {
	URL               string
	SSLCertPath       string
	ConnectionRetries int
	ShouldMigrate     bool
}

type AssetsConfig struct {
	MaxUploadSizeBytes int64
	InitialCategories  []string
}

type Config struct {
	Server                *ServerConfig
	DB                    *DatabaseConfig
	Assets                *AssetsConfig
	SessionTokenSecretKey string
	SessionTokenDuration  time.Duration
	BrevoAPIKey           string
	EmailSenderAddress    string
	EmailLogoURL          string
	ServerFixedToken      string
	NomadAddr             string
	NomadToken            string
	NomadTemplatePath     string
	NomadImageName        string
	FTRServerImage        string
}

func CreateConfig() *Config {
	dbc := &DatabaseConfig{
		URL:               os.Getenv("DATABASE_URL"),
		SSLCertPath:       os.Getenv("DATABASE_SSL_CERT_PATH"),
		ConnectionRetries: getEnvOrDefaultInt("DB_CONNECTION_RETRIES", 10),
		ShouldMigrate:     getEnvOrDefaultString("DB_SHOULD_MIGRATE", "false") == "true",
	}

	assetsConf := &AssetsConfig{
		MaxUploadSizeBytes: int64(getEnvOrDefaultInt("ASSETS_MAX_UPLOAD_SIZE_BYTES", 20*1024*1024)),
		InitialCategories:  getEnvOrDefaultStringList("ASSETS_INITIAL_CATEGORIES", []string{"weapons", "consumables"}),
	}

	serverConf := &ServerConfig{
		Hostname:        getEnvOrDefaultString("SERVER_HOSTNAME", "localhost"),
		Port:            getEnvOrDefaultInt("SERVER_PORT", 8000),
		ShutdownTimeout: getEnvOrDefaultDuration("SERVER_SHUTDOWN_TIMEOUT", time.Second*30),
		Environment:     getEnvironmentType(os.Getenv("SERVER_ENVIRONMENT")),
		PublicIP:        os.Getenv("PUBLIC_IP"),
	}

	return &Config{
		Server:                serverConf,
		DB:                    dbc,
		Assets:                assetsConf,
		SessionTokenSecretKey: os.Getenv("SESSION_TOKEN_SECRET_KEY"),
		SessionTokenDuration:  getEnvOrDefaultDuration("SESSION_TOKEN_DURATION", time.Hour*24),
		BrevoAPIKey:           os.Getenv("BREVO_API_KEY"),
		EmailSenderAddress:    os.Getenv("EMAIL_SENDER_ADDRESS"),
		EmailLogoURL:          getEnvOrDefaultString("EMAIL_LOGO_URL", "https://avatars.githubusercontent.com/u/231922724?s=400&u=5f4eb45fb6dc7cfa42333bfe1dc64a376122e3d0&v=4"),
		ServerFixedToken:      os.Getenv("SERVER_FIXED_TOKEN"),
		NomadAddr:             os.Getenv("NOMAD_ADDR"),
		NomadToken:            os.Getenv("NOMAD_TOKEN"),
		NomadTemplatePath:     getEnvOrDefaultString("NOMAD_TEMPLATE_PATH", "/nomad/templates/ftr-server-job.nomad"),
		FTRServerImage:        os.Getenv("FTR_SERVER_IMAGE"),
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

func getEnvOrDefaultStringList(key string, defaultValue []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}

	return result
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
