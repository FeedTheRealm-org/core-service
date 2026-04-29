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
	Hostname              string
	Port                  int
	ShutdownTimeout       time.Duration
	Environment           EnvironmentType
	AdminEmail            string
	AdminPassword         string
	PublicIP              string
	SubscriptionOn        bool
	CreatorRevenuePercent float64
}

type DatabaseConfig struct {
	URL               string
	SSLCertPath       string
	ConnectionRetries int
	ShouldMigrate     bool
}

type AssetsConfig struct {
	MaxUploadSizeBytes  int64
	CosmeticsBucketName string
	WorldsBucketName    string
}

type StripeConfig struct {
	StripeApiKey                     string
	StripeGemsWebhookSecret          string
	StripeSubscriptionsWebhookSecret string
	StripeZonePrice                  float64
	StripeBillingAnchorDay           int
	StripeBillingTimezone            string
}

type Config struct {
	Server                *ServerConfig
	DB                    *DatabaseConfig
	Assets                *AssetsConfig
	Stripe                *StripeConfig
	SessionTokenSecretKey string
	SessionTokenDuration  time.Duration
	BrevoAPIKey           string
	EmailSenderAddress    string
	EmailLogoURL          string
	ServerFixedToken      string
	NomadAddr             string
	NomadToken            string
	NomadCertPath         string
	NomadTemplatePath     string
	NomadImageName        string
	ConsulAddr            string
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
		MaxUploadSizeBytes:  int64(getEnvOrDefaultInt("ASSETS_MAX_UPLOAD_SIZE_BYTES", 20*1024*1024)),
		CosmeticsBucketName: getEnvOrDefaultString("ASSETS_COSMETICS_BUCKET_NAME", "cosmetics"),
		WorldsBucketName:    getEnvOrDefaultString("ASSETS_WORLDS_BUCKET_NAME", "worlds"),
	}

	serverConf := &ServerConfig{
		Hostname:              getEnvOrDefaultString("SERVER_HOSTNAME", "localhost"),
		Port:                  getEnvOrDefaultInt("SERVER_PORT", 8000),
		ShutdownTimeout:       getEnvOrDefaultDuration("SERVER_SHUTDOWN_TIMEOUT", time.Second*30),
		Environment:           getEnvironmentType(os.Getenv("SERVER_ENVIRONMENT")),
		AdminEmail:            getEnvOrDefaultString("SERVER_ADMIN_EMAIL", ""),
		AdminPassword:         getEnvOrDefaultString("SERVER_ADMIN_PASSWORD", ""),
		PublicIP:              os.Getenv("PUBLIC_IP"),
		SubscriptionOn:        getEnvOrDefaultBool("SUBSCRIPTION_ON", true),
		CreatorRevenuePercent: getEnvOrDefaultFloat("CREATOR_REVENUE_PERCENT", 0.1),
	}

	stripeConf := &StripeConfig{
		StripeApiKey:                     os.Getenv("STRIPE_API_KEY"),
		StripeGemsWebhookSecret:          os.Getenv("STRIPE_GEMS_WEBHOOK_SECRET"),
		StripeSubscriptionsWebhookSecret: os.Getenv("STRIPE_SUBSCRIPTIONS_WEBHOOK_SECRET"),
		StripeZonePrice:                  getEnvOrDefaultFloat("STRIPE_ZONE_PRICE", 5.00),
		StripeBillingAnchorDay:           getEnvOrDefaultInt("STRIPE_BILLING_ANCHOR_DAY", 5),
		StripeBillingTimezone:            getEnvOrDefaultString("STRIPE_BILLING_TIMEZONE", "America/Argentina/Buenos_Aires"),
	}

	return &Config{
		Server:                serverConf,
		DB:                    dbc,
		Assets:                assetsConf,
		Stripe:                stripeConf,
		SessionTokenSecretKey: os.Getenv("SESSION_TOKEN_SECRET_KEY"),
		SessionTokenDuration:  getEnvOrDefaultDuration("SESSION_TOKEN_DURATION", time.Hour*24),
		BrevoAPIKey:           os.Getenv("BREVO_API_KEY"),
		EmailSenderAddress:    os.Getenv("EMAIL_SENDER_ADDRESS"),
		EmailLogoURL:          getEnvOrDefaultString("EMAIL_LOGO_URL", "https://avatars.githubusercontent.com/u/231922724?s=400&u=5f4eb45fb6dc7cfa42333bfe1dc64a376122e3d0&v=4"),
		ServerFixedToken:      os.Getenv("SERVER_FIXED_TOKEN"),
		NomadAddr:             os.Getenv("NOMAD_ADDR"),
		NomadToken:            os.Getenv("NOMAD_TOKEN"),
		NomadCertPath:         os.Getenv("NOMAD_CERT_PATH"),
		NomadTemplatePath:     getEnvOrDefaultString("NOMAD_TEMPLATE_PATH", "/nomad/templates/ftr-server-job.nomad"),
		ConsulAddr:            os.Getenv("CONSUL_ADDR"),
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

func getEnvOrDefaultFloat(key string, defaultValue float64) float64 {
	value, err := strconv.ParseFloat(os.Getenv(key), 64)
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

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
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
