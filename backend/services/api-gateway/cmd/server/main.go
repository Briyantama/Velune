// API gateway (strangler): forwards /api/v1/* to configured upstreams by path prefix, or LEGACY_API_URL.
package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	sharedconfig "github.com/moon-eye/velune/shared/config"
	constx "github.com/moon-eye/velune/shared/constx"
	httpx "github.com/moon-eye/velune/shared/httpx"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
	"github.com/moon-eye/velune/shared/otelx"
	"github.com/moon-eye/velune/shared/stringx"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("api-gateway")
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg, err := sharedconfig.Load("api-gateway")
	if err != nil {
		log.Fatal("config", zap.Error(err))
	}

	if err := otelx.Init(context.Background(), otelx.Options{ServiceName: cfg.ServiceName}); err != nil {
		log.Fatal("otel_init", zap.Error(err))
	}
	defer func() {
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = otelx.Shutdown(sctx)
	}()
	log.Info("tracing_exporter", zap.String("mode", otelx.ExporterMode()))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /api/v1/gateway/routes", routesHandler(cfg))
	mux.Handle("GET /metrics", metrics.Handler())

	mux.HandleFunc("/api/v1/reports", reportProxyHandler(cfg, log))
	mux.HandleFunc("/api/v1/reports/", reportProxyHandler(cfg, log))

	mux.HandleFunc("/api/v1/", catchAllHandler(cfg, log))

	handler := otelx.HTTPHandler(middlewares.CorrelationIDHeader(instrumentGateway(cfg, log, mux)), "http.server")

	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Info("gateway listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, constx.StatusOK, map[string]string{
		"status":  "ok",
		"service": "api-gateway",
	})
}

func instrumentGateway(cfg *sharedconfig.Service, log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grp := classifyRoute(cfg, r.URL.Path)
		metrics.GatewayRequestsTotal.WithLabelValues(grp).Inc()
		fields := append(sharedlog.FieldsFromContext(r.Context()),
			zap.String("route_group", grp),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
		)
		log.Info("gateway_request", fields...)
		next.ServeHTTP(w, r)
	})
}

func classifyRoute(cfg *sharedconfig.Service, path string) string {
	switch {
	case path == "/health" || path == "/api/v1/gateway/routes":
		return "meta"
	case stringx.HasPrefix(path, "/api/v1/reports"):
		return "reports"
	case stringx.HasPrefix(path, "/api/v1/") && pickProxy(cfg, path) != nil:
		return "microservice"
	case stringx.HasPrefix(path, "/api/v1/") && os.Getenv("LEGACY_API_URL") != "":
		return "legacy_catchall"
	default:
		return "unknown"
	}
}

func routesHandler(cfg *sharedconfig.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, constx.StatusOK, map[string]string{
			"auth":         cfg.AuthServiceURL,
			"transaction":  cfg.TransactionServiceURL,
			"category":     cfg.CategoryServiceURL,
			"budget":       cfg.BudgetServiceURL,
			"report":       cfg.ReportServiceURL,
			"notification": cfg.NotificationServiceURL,
			"legacy":       os.Getenv("LEGACY_API_URL"),
		})
	}
}

func reportProxyHandler(cfg *sharedconfig.Service, log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		primary := cfg.ReportServiceURL
		if primary == "" {
			metrics.GatewayFallbackHitsTotal.WithLabelValues("report_no_upstream").Inc()
			writeGatewayError(w, "REPORT_UPSTREAM_UNAVAILABLE", "report-service URL not configured", constx.StatusBadGateway)
			return
		}
		u, err := url.Parse(primary)
		if err != nil {
			writeGatewayError(w, "CONFIG_ERROR", "invalid report-service URL", constx.StatusInternalServerError)
			return
		}
		p := httpx.SingleHostReverseProxy(u)
		p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("report_upstream_error", append(sharedlog.FieldsFromContext(r.Context()),
				zap.Error(err),
				zap.String("upstream", u.Host),
			)...)
			metrics.GatewayFallbackHitsTotal.WithLabelValues("report_upstream_error").Inc()
			writeGatewayError(w, "REPORT_UPSTREAM_FAILED", "report-service unavailable (legacy report fallback retired)", constx.StatusBadGateway)
		}
		p.ServeHTTP(w, r)
	}
}

func catchAllHandler(cfg *sharedconfig.Service, log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h := pickProxy(cfg, r.URL.Path); h != nil {
			h.ServeHTTP(w, r)
			return
		}
		if leg := os.Getenv("LEGACY_API_URL"); leg != "" {
			metrics.GatewayFallbackHitsTotal.WithLabelValues("legacy_catchall").Inc()
			log.Warn("gateway_legacy_catchall", append(sharedlog.FieldsFromContext(r.Context()),
				zap.String("path", r.URL.Path),
			)...)
			httpx.MustProxy(leg).ServeHTTP(w, r)
			return
		}
		httpx.WriteJSON(w, constx.StatusNotFound, nil)
	}
}

func writeGatewayError(w http.ResponseWriter, code, message string, status int) {
	httpx.WriteJSON(w, status, nil)
}

func pickProxy(cfg *sharedconfig.Service, path string) http.Handler {
	type rule struct {
		prefix string
		target string
	}
	rules := []rule{
		{"/api/v1/auth", cfg.AuthServiceURL},
		{"/api/v1/transactions", cfg.TransactionServiceURL},
		{"/api/v1/accounts", cfg.TransactionServiceURL},
		{"/api/v1/recurring", cfg.TransactionServiceURL},
		{"/api/v1/categories", cfg.TransactionServiceURL},
		{"/api/v1/budgets", cfg.BudgetServiceURL},
		{"/api/v1/notifications", cfg.NotificationServiceURL},
	}
	for _, ru := range rules {
		if ru.target == "" {
			continue
		}
		if stringx.HasPrefix(path, ru.prefix) {
			return httpx.MustProxy(ru.target)
		}
	}
	return nil
}
