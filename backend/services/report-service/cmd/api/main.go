package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	httpapi "github.com/moon-eye/velune/services/report-service/internal/delivery/http"
	"github.com/moon-eye/velune/services/report-service/internal/infrastructure/transactions"
	"github.com/moon-eye/velune/services/report-service/internal/usecase"
	sharedconfig "github.com/moon-eye/velune/shared/config"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("report-service")
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg, err := sharedconfig.Load("report-service")
	if err != nil {
		log.Fatal("config", zap.Error(err))
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	txClient := transactions.New(cfg.TransactionServiceURL)
	reportSvc := &usecase.ReportService{Transactions: txClient}
	server := &httpapi.Server{
		Reports:   reportSvc,
		Validate:  validator.New(),
		Log:       log,
		JWTSecret: cfg.JWTSecret,
	}
	handler := httpapi.NewRouter(server)

	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("server listening", zap.String("addr", addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
