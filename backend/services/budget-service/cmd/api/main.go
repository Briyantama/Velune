package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	config "github.com/moon-eye/velune/shared/config"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	httpapi "github.com/moon-eye/velune/services/budget-service/internal/delivery/http"
	"github.com/moon-eye/velune/services/budget-service/internal/infrastructure/postgres"
	txclient "github.com/moon-eye/velune/services/budget-service/internal/infrastructure/transactions"
	"github.com/moon-eye/velune/services/budget-service/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("budget-service")
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg, err := config.Load("budget-service")
	if err != nil {
		log.Fatal("config", zap.Error(err))
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	ctx := context.Background()
	if err := runMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatal("migrations", zap.Error(err))
	}

	store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database", zap.Error(err))
	}
	defer store.Close()

	budgetRepo := postgres.NewBudgetRepo(store)
	txSummaryClient := txclient.New(cfg.TransactionServiceURL)

	srv := &httpapi.Server{
		Budgets: &usecase.BudgetService{
			Budgets:      budgetRepo,
			Transactions: txSummaryClient,
		},
		Validate:  validator.New(),
		Log:       log,
		JWTSecret: cfg.JWTSecret,
		DB:        store.Pool,
	}

	handler := httpapi.NewRouter(srv)
	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
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
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", zap.Error(err))
	}
}

func runMigrations(databaseURL, sourceURL string) error {
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

