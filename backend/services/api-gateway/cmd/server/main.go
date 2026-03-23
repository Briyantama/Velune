// API gateway (strangler): forwards /api/v1/* to configured upstreams by path prefix, or LEGACY_API_URL.
package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	sharedconfig "github.com/moon-eye/velune/shared/config"
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
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		if h := pickProxy(cfg, r.URL.Path); h != nil {
			h.ServeHTTP(w, r)
			return
		}
		if leg := os.Getenv("LEGACY_API_URL"); leg != "" {
			mustProxy(leg).ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
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
		_, _ = w.Write([]byte(`{"auth":"` + cfg.AuthServiceURL + `","transaction":"` + cfg.TransactionServiceURL + `","category":"` + cfg.CategoryServiceURL + `","budget":"` + cfg.BudgetServiceURL + `","report":"` + cfg.ReportServiceURL + `","legacy":"` + os.Getenv("LEGACY_API_URL") + `"}`))
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
		{"/api/v1/reports", cfg.ReportServiceURL},
	}
	for _, ru := range rules {
		if ru.target == "" {
			continue
		}
		if stringx.HasPrefix(path, ru.prefix) {
			return mustProxy(ru.target)
		}
	}
	return nil
}

func mustProxy(origin string) http.Handler {
	u, err := url.Parse(origin)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad upstream URL", http.StatusInternalServerError)
		})
	}
	return httputil.NewSingleHostReverseProxy(u)
}
