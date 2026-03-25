package otelx

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPHandler wraps an http.Handler with HTTP server tracing (works with chi and plain mux).
func HTTPHandler(handler http.Handler, operation string) http.Handler {
	if operation == "" {
		operation = "http.server"
	}
	return otelhttp.NewHandler(handler, operation,
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			if r != nil && r.Pattern != "" {
				return r.Method + " " + r.Pattern
			}
			return operation
		}),
	)
}

// TracedHTTPClient returns a copy of base with the Transport wrapped for client spans, or a new client if base is nil.
func TracedHTTPClient(base *http.Client) *http.Client {
	if base == nil {
		base = &http.Client{Timeout: 30 * time.Second}
	}
	t := base.Transport
	if t == nil {
		t = http.DefaultTransport
	}
	base = &http.Client{
		Transport:     otelhttp.NewTransport(t),
		CheckRedirect: base.CheckRedirect,
		Jar:           base.Jar,
		Timeout:       base.Timeout,
	}
	return base
}

// InjectHTTP writes the current span context into headers (e.g. reverse proxy Director).
func InjectHTTP(ctx context.Context, hdr http.Header) {
	if hdr == nil {
		return
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(hdr))
}
