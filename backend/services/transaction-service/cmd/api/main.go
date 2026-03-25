package main

import (
	"context"
	"encoding/json"
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
	"github.com/google/uuid"
	httpapi "github.com/moon-eye/velune/services/transaction-service/internal/delivery/http"
	"github.com/moon-eye/velune/services/transaction-service/internal/infrastructure/postgres"
	"github.com/moon-eye/velune/services/transaction-service/internal/usecase"
	config "github.com/moon-eye/velune/shared/config"
	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/events"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("transaction-service")
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg, err := config.Load("transaction-service")
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

	accountRepo := postgres.NewAccountRepo(store)
	categoryRepo := postgres.NewCategoryRepo(store)
	txRepo := postgres.NewTransactionRepo(store)
	recurringRepo := postgres.NewRecurringRepo(store)
	ledger := postgres.NewLedger(store)
	eventPublisher, err := events.NewRabbitPublisher(cfg.BrokerURL, cfg.BrokerExchange, cfg.BrokerRoutingKey, cfg.BrokerDLX, cfg.BrokerDLQRoutingKey)
	if err != nil {
		log.Fatal("events", zap.Error(err))
	}
	defer eventPublisher.Close()

	v := validator.New()
	srv := &httpapi.Server{
		Accounts:     &usecase.AccountService{Accounts: accountRepo},
		Categories:   &usecase.CategoryService{Categories: categoryRepo},
		Transactions: &usecase.TransactionService{Ledger: ledger, Transactions: txRepo, Accounts: accountRepo, Categories: categoryRepo, Logger: log},
		Recurring:    &usecase.RecurringService{Recurring: recurringRepo, Accounts: accountRepo, Categories: categoryRepo},
		Validate:     v,
		Log:          log,
		JWTSecret:    cfg.JWTSecret,
		DB:           store.Pool,
	}

	handler := httpapi.NewRouter(srv)
	addr := ":" + cfg.HTTPPort
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := dispatchOutboxBatch(ctx, store, eventPublisher, cfg.OutboxBatchSize, cfg.OutboxMaxRetry, cfg.RetryBaseDelaySeconds, log); err != nil {
					log.Error("outbox dispatch failed", zap.Error(err))
				}
			}
		}
	}()

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

func dispatchOutboxBatch(
	ctx context.Context,
	store *postgres.Store,
	publisher *events.RabbitPublisher,
	batchSize, maxRetry, baseDelaySeconds int,
	log *zap.Logger,
) error {
	if batchSize <= 0 {
		batchSize = 50
	}
	rows, err := store.Pool.Query(ctx, `
		SELECT id, payload, retry_count
		FROM event_outbox
		WHERE status IN ('pending','failed')
		  AND retry_count < $1
		  AND next_retry_at <= now()
		ORDER BY created_at ASC
		LIMIT $2
	`, maxRetry, batchSize)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id [16]byte
		var payload []byte
		var retryCount int
		if err := rows.Scan(&id, &payload, &retryCount); err != nil {
			return err
		}
		var env contracts.EventEnvelope
		if err := json.Unmarshal(payload, &env); err != nil {
			_, _ = store.Pool.Exec(ctx, `UPDATE event_outbox SET status='failed', retry_count=retry_count+1, updated_at=now() WHERE id=$1`, id)
			continue
		}
		if env.EventID == (contracts.EventEnvelope{}).EventID {
			env.EventID = uuidFromBytes(id)
		}
		if env.Idempotency == "" {
			env.Idempotency = "outbox:" + env.EventType + ":" + env.EventID.String()
		}
		if err := publisher.Publish(ctx, env); err != nil {
					if retryCount+1 >= maxRetry {
						_ = publisher.PublishDLQ(ctx, env)
						_, _ = store.Pool.Exec(ctx, `UPDATE event_outbox SET status='failed', retry_count=retry_count+1, updated_at=now() WHERE id=$1`, id)
						log.Error("outbox moved to dlq", zap.String("event_id", env.EventID.String()))
						continue
					}
					nextRetry := time.Now().UTC().Add(time.Duration(1<<retryCount*maxInt(1, baseDelaySeconds)) * time.Second)
			_, _ = store.Pool.Exec(ctx, `
				UPDATE event_outbox
				SET status='failed', retry_count=retry_count+1, next_retry_at=$2, updated_at=now()
				WHERE id=$1
			`, id, nextRetry)
			log.Warn("outbox publish retry", zap.String("event_id", env.EventID.String()), zap.Int("retry_count", retryCount+1))
			continue
		}
		_, _ = store.Pool.Exec(ctx, `UPDATE event_outbox SET status='sent', updated_at=now() WHERE id=$1`, id)
		log.Info("outbox published", zap.String("event_id", env.EventID.String()), zap.String("event_type", env.EventType))
	}
	return rows.Err()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func uuidFromBytes(b [16]byte) uuid.UUID {
	return uuid.UUID(b)
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
