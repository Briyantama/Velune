package otelx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	shutdownFn   func(context.Context) error
	shutdownMu   sync.Mutex
	propagatorMu sync.Once
)

// Options configures the global TracerProvider. Zero value is valid; ServiceName defaults from OTEL_SERVICE_NAME.
type Options struct {
	ServiceName string
}

// IsTracingEnabled returns true when traces should be exported (stdout or OTLP), per env convention in backend docs.
func IsTracingEnabled() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_TRACES_ENABLED")), "true") {
		return true
	}
	if ep := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")); ep != "" {
		return true
	}
	if ep := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")); ep != "" {
		return true
	}
	return false
}

// ExporterMode is stdout, otlp, or noop when tracing is disabled.
func ExporterMode() string {
	if !IsTracingEnabled() {
		return "noop"
	}
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != "" ||
		strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")) != "" {
		return "otlp"
	}
	m := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_TRACES_EXPORTER")))
	if m == "otlp" {
		return "otlp"
	}
	return "stdout"
}

// Init installs W3C tracecontext + baggage propagation and, when enabled, a real TracerProvider.
// Call Shutdown from main with defer. Safe to call multiple times; only the first successful init registers shutdown.
func Init(ctx context.Context, opts Options) error {
	propagatorMu.Do(func() {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	})

	if !IsTracingEnabled() {
		return nil
	}

	serviceName := strings.TrimSpace(opts.ServiceName)
	if serviceName == "" {
		serviceName = strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	}
	if serviceName == "" {
		serviceName = "velune"
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return fmt.Errorf("otel resource: %w", err)
	}

	var exp sdktrace.SpanExporter
	switch ExporterMode() {
	case "otlp":
		exp, err = otlptracehttp.New(ctx)
		if err != nil {
			return fmt.Errorf("otlp exporter: %w", err)
		}
	default:
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return fmt.Errorf("stdout exporter: %w", err)
		}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	shutdownMu.Lock()
	shutdownFn = tp.Shutdown
	shutdownMu.Unlock()
	return nil
}

// Shutdown flushes and shuts down the SDK TracerProvider if Init configured one.
func Shutdown(ctx context.Context) error {
	shutdownMu.Lock()
	fn := shutdownFn
	shutdownFn = nil
	shutdownMu.Unlock()
	if fn == nil {
		return nil
	}
	return fn(ctx)
}
