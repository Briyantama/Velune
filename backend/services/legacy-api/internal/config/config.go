package config

import (
	"os"
	"time"

	"github.com/moon-eye/velune/shared/stringx"
	"github.com/spf13/viper"
)

type Config struct {
	HTTPPort         string
	DatabaseURL      string
	JWTSecret        string
	JWTExpiry        time.Duration
	MigrationsPath   string
	Environment      string
}

func Load() (*Config, error) {
	viper.SetEnvKeyReplacer(stringx.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("MIGRATIONS_PATH", "file://migrations")
	viper.SetDefault("ENV", "development")

	jwtExp, err := time.ParseDuration(viper.GetString("JWT_EXPIRY"))
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	c := &Config{
		HTTPPort:       viper.GetString("HTTP_PORT"),
		DatabaseURL:    viper.GetString("DATABASE_URL"),
		JWTSecret:      viper.GetString("JWT_SECRET"),
		JWTExpiry:      jwtExp,
		MigrationsPath: viper.GetString("MIGRATIONS_PATH"),
		Environment:    viper.GetString("ENV"),
	}
	if c.DatabaseURL == "" {
		c.DatabaseURL = os.Getenv("DATABASE_URL")
	}
	if c.JWTSecret == "" {
		c.JWTSecret = os.Getenv("JWT_SECRET")
	}
	if c.MigrationsPath == "" {
		c.MigrationsPath = os.Getenv("MIGRATIONS_PATH")
		if c.MigrationsPath == "" {
			c.MigrationsPath = "file://migrations"
		}
	}
	return c, nil
}
