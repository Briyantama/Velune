package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a production logger by default; use LOG_FORMAT=dev for console development output.
func New(serviceName string) (*zap.Logger, error) {
	if os.Getenv("LOG_FORMAT") == "dev" {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		l, err := cfg.Build(zap.Fields(zap.String("service", serviceName)))
		return l, err
	}
	cfg := zap.NewProductionConfig()
	cfg.InitialFields = map[string]interface{}{"service": serviceName}
	return cfg.Build()
}
