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
	httpapi "github.com/moon-eye/velune/services/transaction-service/internal/delivery/http"
	"github.com/moon-eye/velune/services/transaction-service/internal/infrastructure/postgres"
	tsrecon "github.com/moon-eye/velune/services/transaction-service/internal/reconciliation"
	"github.com/moon-eye/velune/services/transaction-service/internal/usecase"
	config "github.com/moon-eye/velune/shared/config"
	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/events"
	"github.com/moon-eye/velune/shared/helper"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/otelx"
	"github.com/moon-eye/velune/shared/sim"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
	stringx "github.com/moon-eye/velune/shared/stringx"
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

	if err := otelx.Init(context.Background(), otelx.Options{ServiceName: cfg.ServiceName}); err != nil {
		log.Fatal("otel_init", zap.Error(err))
	}
	defer func() {
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = otelx.Shutdown(sctx)
	}()
	log.Info("tracing_exporter", zap.String("mode", otelx.ExporterMode()))

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
	chaos := sim.LoadFromEnv()
	eventPublisher, err := events.NewRabbitPublisher(cfg.BrokerURL, cfg.BrokerExchange, cfg.BrokerRoutingKey, cfg.BrokerDLX, cfg.BrokerDLQRoutingKey, chaos)
	if err != nil {
		log.Fatal("events", zap.Error(err))
	}
	defer eventPublisher.Close()

	v := validator.New()
	srv := &httpapi.Server{
		Accounts:         &usecase.AccountService{Accounts: accountRepo},
		Categories:       &usecase.CategoryService{Categories: categoryRepo},
		Transactions:     &usecase.TransactionService{Ledger: ledger, Transactions: txRepo, Accounts: accountRepo, Categories: categoryRepo, Logger: log},
		Recurring:        &usecase.RecurringService{Recurring: recurringRepo, Accounts: accountRepo, Categories: categoryRepo},
		Validate:         v,
		Log:              log,
		JWTSecret:        cfg.JWTSecret,
		AdminInternalKey: stringx.TrimSpace(os.Getenv("ADMIN_INTERNAL_KEY")),
		DB:               store.Pool,
	}

	handler := otelx.HTTPHandler(httpapi.NewRouter(srv), "http.server")
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
				if err := dispatchOutboxBatch(ctx, store, eventPublisher, "transaction-service", cfg.OutboxBatchSize, cfg.OutboxMaxRetry, cfg.RetryBaseDelaySeconds, log); err != nil {
					log.Error("outbox dispatch failed", append(sharedlog.FieldsFromContext(ctx), zap.Error(err))...)
				}
			}
		}
	}()

	if cfg.ReconcileInterval > 0 {
		go func() {
			ticker := time.NewTicker(cfg.ReconcileInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					rctx := context.Background()
					if _, _, err := tsrecon.ReconcileAccountBalances(rctx, store.Pool, log); err != nil {
						log.Error("balance_reconcile", zap.Error(err))
					}
				}
			}
		}()
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

func dispatchOutboxBatch(
	ctx context.Context,
	store *postgres.Store,
	publisher *events.RabbitPublisher,
	serviceName string,
	batchSize, maxRetry, baseDelaySeconds int,
	log *zap.Logger,
) error {
	if batchSize <= 0 {
		batchSize = 50
	}
	metrics.RefreshOutboxPending(ctx, store.Pool, maxRetry)
	q := db.New(store.Pool)
	items, err := q.OutboxListDueForDispatch(ctx, db.OutboxListDueForDispatchParams{
		RetryCount: int32(maxRetry),
		Limit:      int32(batchSize),
	})
	if err != nil {
		return err
	}

	for _, row := range items {
		id := row.ID
		payload := row.Payload
		retryCount := int(row.RetryCount)
		var env contracts.EventEnvelope
		if err := json.Unmarshal(payload, &env); err != nil {
			_ = q.OutboxMarkFailedBumpRetry(ctx, id)
			continue
		}
		if env.EventID == (contracts.EventEnvelope{}).EventID {
			env.EventID = helper.FromPgUUID(id)
		}
		outboxID := helper.FromPgUUID(id)
		if env.Idempotency == "" {
			env.Idempotency = "outbox:" + env.EventType + ":" + env.EventID.String()
		}
		log.Info("outbox_publish_attempt", append(sharedlog.FieldsFromContext(ctx),
			zap.String("outbox_id", outboxID.String()),
			zap.String("event_id", env.EventID.String()),
			zap.String("event_type", env.EventType),
			zap.Int("retry_count", retryCount),
		)...)
		if err := publisher.Publish(ctx, env); err != nil {
			if retryCount+1 >= maxRetry {
				_ = publisher.PublishDLQ(ctx, env)
				metrics.DLQMessagesTotal.WithLabelValues(serviceName).Inc()
				_ = q.OutboxMarkFailedBumpRetry(ctx, id)
				log.Error("outbox_publish_dlq", append(sharedlog.FieldsFromContext(ctx),
					zap.String("outbox_id", outboxID.String()),
					zap.String("event_id", env.EventID.String()),
					zap.String("event_type", env.EventType),
					zap.Int("retry_count", retryCount+1),
				)...)
				continue
			}
			metrics.OutboxRetryTotal.WithLabelValues(serviceName).Inc()
			nextRetry := time.Now().UTC().Add(time.Duration(1<<retryCount*helper.MaxInt(1, baseDelaySeconds)) * time.Second)
			_ = q.OutboxMarkFailedScheduleRetry(ctx, db.OutboxMarkFailedScheduleRetryParams{
				ID:          id,
				NextRetryAt: helper.ToPgTS(nextRetry),
			})
			log.Warn("outbox_publish_retry_scheduled", append(sharedlog.FieldsFromContext(ctx),
				zap.String("outbox_id", outboxID.String()),
				zap.String("event_id", env.EventID.String()),
				zap.String("event_type", env.EventType),
				zap.Int("retry_count", retryCount+1),
				zap.Time("next_retry_at", nextRetry),
				zap.Error(err),
			)...)
			continue
		}
		_ = q.OutboxMarkSent(ctx, id)
		log.Info("outbox_publish_sent", append(sharedlog.FieldsFromContext(ctx),
			zap.String("outbox_id", outboxID.String()),
			zap.String("event_id", env.EventID.String()),
			zap.String("event_type", env.EventType),
		)...)
	}
	return nil
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
