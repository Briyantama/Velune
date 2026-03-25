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
	AuthServiceURL         string
	TransactionServiceURL  string
	CategoryServiceURL     string
	BudgetServiceURL       string
	ReportServiceURL       string
	NotificationServiceURL string
	BrokerURL              string
	BrokerExchange         string
	BrokerQueue            string
	BrokerRoutingKey       string
	EmailFrom              string
	OutboxBatchSize        int
	OutboxMaxRetry         int
	RetryBaseDelaySeconds  int
	BrokerDLX              string
	BrokerDLQ              string
	BrokerDLQRoutingKey    string
	// ReconcileInterval runs periodic reconciliation when >0 (e.g. 1h, 24h). Empty/disabled by default.
	ReconcileInterval time.Duration
}

func Load(serviceName string) (*Service, error) {
	viper.SetEnvKeyReplacer(stringx.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("HTTP_HOST", "0.0.0.0")
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("MIGRATIONS_PATH", "file://migrations")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("BROKER_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("BROKER_EXCHANGE", "velune.events")
	viper.SetDefault("BROKER_QUEUE", "notification.overspend")
	viper.SetDefault("BROKER_ROUTING_KEY", "notification.overspend.requested")
	viper.SetDefault("EMAIL_FROM", "noreply@velune.local")
	viper.SetDefault("OUTBOX_BATCH_SIZE", 50)
	viper.SetDefault("OUTBOX_MAX_RETRY", 7)
	viper.SetDefault("RETRY_BASE_DELAY_SECONDS", 2)
	viper.SetDefault("BROKER_DLX", "velune.events.dlx")
	viper.SetDefault("BROKER_DLQ", "velune.events.dlq")
	viper.SetDefault("BROKER_DLQ_ROUTING_KEY", "velune.dlq")
	viper.SetDefault("RECONCILE_INTERVAL", "")

	jwtExp, err := time.ParseDuration(viper.GetString("JWT_EXPIRY"))
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	reconcileDur := time.Duration(0)
	if s := viper.GetString("RECONCILE_INTERVAL"); s != "" {
		d, err := time.ParseDuration(s)
		if err == nil && d > 0 {
			reconcileDur = d
		}
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
		BrokerURL:              viper.GetString("BROKER_URL"),
		BrokerExchange:         viper.GetString("BROKER_EXCHANGE"),
		BrokerQueue:            viper.GetString("BROKER_QUEUE"),
		BrokerRoutingKey:       viper.GetString("BROKER_ROUTING_KEY"),
		EmailFrom:              viper.GetString("EMAIL_FROM"),
		OutboxBatchSize:        viper.GetInt("OUTBOX_BATCH_SIZE"),
		OutboxMaxRetry:         viper.GetInt("OUTBOX_MAX_RETRY"),
		RetryBaseDelaySeconds:  viper.GetInt("RETRY_BASE_DELAY_SECONDS"),
		BrokerDLX:              viper.GetString("BROKER_DLX"),
		BrokerDLQ:              viper.GetString("BROKER_DLQ"),
		BrokerDLQRoutingKey:    viper.GetString("BROKER_DLQ_ROUTING_KEY"),
		ReconcileInterval:      reconcileDur,
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
