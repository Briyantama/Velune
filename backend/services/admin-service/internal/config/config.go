package config

import (
	"os"
	"strconv"

	"github.com/moon-eye/velune/shared/stringx"
)

// Config holds admin-service-only settings (env-driven).
type Config struct {
	HTTPHost string
	HTTPPort string

	AdminAPIKey        string
	AdminInternalKey   string
	TransactionDBURL   string
	BudgetDBURL        string
	NotificationDBURL  string
	BrokerURL          string
	BrokerExchange     string
	BrokerDLX          string
	BrokerDLQ           string
	BrokerRoutingKey    string
	BrokerDLQRoutingKey string

	AuthServiceURL          string
	TransactionServiceURL   string
	CategoryServiceURL      string
	BudgetServiceURL        string
	ReportServiceURL        string
	NotificationServiceURL  string
}

func Load() *Config {
	c := &Config{
		HTTPHost:               firstNonEmpty(os.Getenv("HTTP_HOST"), "0.0.0.0"),
		HTTPPort:               firstNonEmpty(os.Getenv("HTTP_PORT"), "8099"),
		AdminAPIKey:            stringx.TrimSpace(os.Getenv("ADMIN_API_KEY")),
		AdminInternalKey:       stringx.TrimSpace(os.Getenv("ADMIN_INTERNAL_KEY")),
		TransactionDBURL:       os.Getenv("TRANSACTION_DATABASE_URL"),
		BudgetDBURL:            os.Getenv("BUDGET_DATABASE_URL"),
		NotificationDBURL:      os.Getenv("NOTIFICATION_DATABASE_URL"),
		BrokerURL:              firstNonEmpty(os.Getenv("BROKER_URL"), "amqp://guest:guest@localhost:5672/"),
		BrokerExchange:         firstNonEmpty(os.Getenv("BROKER_EXCHANGE"), "velune.events"),
		BrokerDLX:              os.Getenv("BROKER_DLX"),
		BrokerDLQ:              os.Getenv("BROKER_DLQ"),
		BrokerRoutingKey:       firstNonEmpty(os.Getenv("BROKER_ROUTING_KEY"), "notification.overspend.requested"),
		BrokerDLQRoutingKey:    firstNonEmpty(os.Getenv("BROKER_DLQ_ROUTING_KEY"), "velune.dlq"),
		AuthServiceURL:         os.Getenv("AUTH_SERVICE_URL"),
		TransactionServiceURL:  os.Getenv("TRANSACTION_SERVICE_URL"),
		CategoryServiceURL:     os.Getenv("CATEGORY_SERVICE_URL"),
		BudgetServiceURL:       os.Getenv("BUDGET_SERVICE_URL"),
		ReportServiceURL:       os.Getenv("REPORT_SERVICE_URL"),
		NotificationServiceURL: os.Getenv("NOTIFICATION_SERVICE_URL"),
	}
	if c.TransactionDBURL == "" {
		c.TransactionDBURL = os.Getenv("DATABASE_URL_TRANSACTION")
	}
	if c.BudgetDBURL == "" {
		c.BudgetDBURL = os.Getenv("DATABASE_URL_BUDGET")
	}
	if c.NotificationDBURL == "" {
		c.NotificationDBURL = os.Getenv("DATABASE_URL_NOTIFICATION")
	}
	return c
}

func firstNonEmpty(a, b string) string {
	if stringx.TrimSpace(a) != "" {
		return a
	}
	return b
}

// AtoiDefault parses s as int or returns def.
func AtoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
