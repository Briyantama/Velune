package otelx

import (
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InstrumentPoolConfig attaches an OpenTelemetry tracer to pgx when tracing is enabled.
func InstrumentPoolConfig(cfg *pgxpool.Config) {
	if cfg == nil || !IsTracingEnabled() {
		return
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()
}
