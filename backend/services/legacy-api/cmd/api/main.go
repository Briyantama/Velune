package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpx "github.com/moon-eye/velune/services/legacy-api/internal/delivery/http"
	sharedconfig "github.com/moon-eye/velune/shared/config"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("legacy-api")
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg, err := sharedconfig.Load("legacy-api")
	if err != nil {
		log.Fatal("config", zap.Error(err))
	}

	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	if cfg.HTTPPort == "" {
		addr = "0.0.0.0:8090"
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           httpx.NewRouter(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("legacy-api shell listening", zap.String("addr", addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", zap.Error(err))
	}
}
