package config

import (
	"os"
	"time"

	"github.com/moon-eye/velune/shared/stringx"
	"github.com/spf13/viper"
)

// Service holds common environment-driven settings for all deployable services.
type Service struct {
	ServiceName    string
	HTTPPort       string
	HTTPHost       string
	DatabaseURL    string
	Environment    string
	MigrationsPath string
	JWTSecret      string
	JWTExpiry      time.Duration
	// Downstream URLs for clients (report-service, gateway, etc.) — never hardcode in code paths.
	AuthServiceURL        string
	TransactionServiceURL string
	CategoryServiceURL    string
	BudgetServiceURL      string
	ReportServiceURL      string
	NotificationServiceURL string
}

func Load(serviceName string) (*Service, error) {
	viper.SetEnvKeyReplacer(stringx.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("HTTP_HOST", "0.0.0.0")
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("MIGRATIONS_PATH", "file://migrations")
	viper.SetDefault("ENV", "development")

	jwtExp, err := time.ParseDuration(viper.GetString("JWT_EXPIRY"))
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	c := &Service{
		ServiceName:            serviceName,
		HTTPPort:               viper.GetString("HTTP_PORT"),
		HTTPHost:               viper.GetString("HTTP_HOST"),
		DatabaseURL:            viper.GetString("DATABASE_URL"),
		Environment:            viper.GetString("ENV"),
		MigrationsPath:         viper.GetString("MIGRATIONS_PATH"),
		JWTSecret:              viper.GetString("JWT_SECRET"),
		JWTExpiry:              jwtExp,
		AuthServiceURL:         viper.GetString("AUTH_SERVICE_URL"),
		TransactionServiceURL:  viper.GetString("TRANSACTION_SERVICE_URL"),
		CategoryServiceURL:     viper.GetString("CATEGORY_SERVICE_URL"),
		BudgetServiceURL:       viper.GetString("BUDGET_SERVICE_URL"),
		ReportServiceURL:       viper.GetString("REPORT_SERVICE_URL"),
		NotificationServiceURL: viper.GetString("NOTIFICATION_SERVICE_URL"),
	}
	if c.DatabaseURL == "" {
		c.DatabaseURL = os.Getenv("DATABASE_URL")
	}
	if c.JWTSecret == "" {
		c.JWTSecret = os.Getenv("JWT_SECRET")
	}
	if c.MigrationsPath == "" {
		c.MigrationsPath = firstNonEmpty(os.Getenv("MIGRATIONS_PATH"), "file://migrations")
	}
	return c, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
