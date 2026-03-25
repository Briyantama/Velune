// Category service: hierarchical categories and defaults per user.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	sharedconfig "github.com/moon-eye/velune/shared/config"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
	"github.com/moon-eye/velune/shared/otelx"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("category-service")
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	cfg, err := sharedconfig.Load("category-service")
	if err != nil {
		log.Fatal("config", zap.Error(err))
	}
	r := chi.NewRouter()
	r.Use(middlewares.CorrelationIDHeader)
	r.Use(middleware.RealIP, middleware.Recoverer)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"category-service"}`))
	})
	r.Handle("/metrics", metrics.Handler())
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/meta", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"service":"category-service","version":"0.1.0"}`))
		})
	})
	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	srv := &http.Server{Addr: addr, Handler: otelx.HTTPHandler(r, "http.server"), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Info("listening", zap.String("addr", addr))
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
