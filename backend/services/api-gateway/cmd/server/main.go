// API gateway (strangler): forwards /api/v1/* to configured upstreams by path prefix, or LEGACY_API_URL.
package main

import (
	"context"
	"errors"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"api-gateway"}`))
	})
	mux.HandleFunc("GET /api/v1/gateway/routes", routesHandler(cfg))
	mux.HandleFunc("/api/v1/reports", func(w http.ResponseWriter, r *http.Request) {
		primary := cfg.ReportServiceURL
		fallback := os.Getenv("LEGACY_API_URL")
		if primary == "" {
			if fallback != "" {
				httpx.MustProxy(fallback).ServeHTTP(w, r)
				return
			}
		  http.NotFound(w, r)
			return
		}
		reportsProxyWithFallback(primary, fallback).ServeHTTP(w, r)
	})
	mux.HandleFunc("/api/v1/reports/", func(w http.ResponseWriter, r *http.Request) {
		primary := cfg.ReportServiceURL
		fallback := os.Getenv("LEGACY_API_URL")
		if primary == "" {
			if fallback != "" {
				httpx.MustProxy(fallback).ServeHTTP(w, r)
				return
			}
			http.NotFound(w, r)
			return
		}
		reportsProxyWithFallback(primary, fallback).ServeHTTP(w, r)
	})
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		if h := pickProxy(cfg, r.URL.Path); h != nil {
			h.ServeHTTP(w, r)
			return
		}
		if leg := os.Getenv("LEGACY_API_URL"); leg != "" {
			httpx.MustProxy(leg).ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(constx.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"NOT_FOUND","message":"no upstream configured for path"}`))
	})

	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
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

func routesHandler(cfg *sharedconfig.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"auth":"` + cfg.AuthServiceURL + `","transaction":"` + cfg.TransactionServiceURL + `","category":"` + cfg.CategoryServiceURL + `","budget":"` + cfg.BudgetServiceURL + `","report":"` + cfg.ReportServiceURL + `","notification":"` + cfg.NotificationServiceURL + `","legacy":"` + os.Getenv("LEGACY_API_URL") + `"}`))
	}
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

func reportsProxyWithFallback(primaryURL, fallbackURL string) http.Handler {
	primary, err := url.Parse(primaryURL)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad report-service URL", constx.StatusInternalServerError)
		})
	}
	primaryProxy := httpx.SingleHostReverseProxy(primary)
	var fallbackProxy http.Handler
	if fallbackURL != "" {
		fallbackProxy = httpx.MustProxy(fallbackURL)
	}
	primaryProxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == constx.StatusNotFound || resp.StatusCode >= constx.StatusInternalServerError {
			return errors.New("report upstream fallback trigger")
		}
		return nil
	}
	primaryProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, _ error) {
		if fallbackProxy != nil {
			fallbackProxy.ServeHTTP(w, r)
			return
		}
		http.Error(w, "report upstream unavailable", constx.StatusBadGateway)
	}
	return primaryProxy
}
