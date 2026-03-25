package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/moon-eye/velune/services/admin-service/internal/config"
	httpapi "github.com/moon-eye/velune/services/admin-service/internal/delivery/http"
	"github.com/moon-eye/velune/shared/events"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("admin-service")
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg := config.Load()
	if cfg.AdminAPIKey == "" {
		log.Fatal("ADMIN_API_KEY is required for admin-service")
	}

	ctx := context.Background()
	var txPool, bdPool, nfPool *pgxpool.Pool
	if cfg.TransactionDBURL != "" {
		txPool, err = pgxpool.New(ctx, cfg.TransactionDBURL)
		if err != nil {
			log.Fatal("transaction db", zap.Error(err))
		}
		defer txPool.Close()
	}
	if cfg.BudgetDBURL != "" {
		bdPool, err = pgxpool.New(ctx, cfg.BudgetDBURL)
		if err != nil {
			log.Fatal("budget db", zap.Error(err))
		}
		defer bdPool.Close()
	}
	if cfg.NotificationDBURL != "" {
		nfPool, err = pgxpool.New(ctx, cfg.NotificationDBURL)
		if err != nil {
			log.Fatal("notification db", zap.Error(err))
		}
		defer nfPool.Close()
	}

	var pub *events.RabbitPublisher
	if cfg.BrokerURL != "" && cfg.BrokerExchange != "" {
		p, err := events.NewRabbitPublisher(cfg.BrokerURL, cfg.BrokerExchange, cfg.BrokerRoutingKey, cfg.BrokerDLX, cfg.BrokerDLQRoutingKey, nil)
		if err != nil {
			log.Warn("rabbit publisher disabled", zap.Error(err))
		} else {
			pub = p
			defer pub.Close()
		}
	}

	h := httpapi.NewHandlers(cfg, log, txPool, bdPool, nfPool, pub)
	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	srv := &http.Server{
		Addr:              addr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
	}

	go func() {
		log.Info("admin-service listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	shCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shCtx)
}
