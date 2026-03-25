package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	httpapi "github.com/moon-eye/velune/services/notification-service/internal/delivery/http"
	"github.com/moon-eye/velune/services/notification-service/internal/infrastructure/broker"
	"github.com/moon-eye/velune/services/notification-service/internal/infrastructure/delivery"
	"github.com/moon-eye/velune/services/notification-service/internal/infrastructure/email"
	"github.com/moon-eye/velune/services/notification-service/internal/infrastructure/postgres"
	"github.com/moon-eye/velune/services/notification-service/internal/repository"
	"github.com/moon-eye/velune/services/notification-service/internal/usecase"
	config "github.com/moon-eye/velune/shared/config"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/sim"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("notification-service")
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg, err := config.Load("notification-service")
	if err != nil {
		log.Fatal("config", zap.Error(err))
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if err := runMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatal("migrations", zap.Error(err))
	}

	store, err := postgres.NewStore(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database", zap.Error(err))
	}
	defer store.Close()

	chaos := sim.LoadFromEnv()
	rmq, err := broker.New(
		cfg.BrokerURL,
		cfg.BrokerExchange,
		cfg.BrokerQueue,
		cfg.BrokerRoutingKey,
		cfg.BrokerDLX,
		cfg.BrokerDLQ,
		cfg.BrokerDLQRoutingKey,
		chaos,
		log,
	)
	if err != nil {
		log.Fatal("broker", zap.Error(err))
	}
	defer rmq.Close()

	emailStub := &email.StubSender{Log: log, From: cfg.EmailFrom}
	overspend := &usecase.OverspendService{
		Dedupe:    postgres.NewDedupeRepo(store),
		Jobs:      postgres.NewJobRepo(store),
		InApp:     &delivery.InAppChannel{Log: log},
		Email:     &email.ChaosSender{Inner: emailStub, Sim: chaos, Log: log},
		Publisher: rmq,
		MaxRetry:  cfg.OutboxMaxRetry,
		BaseDelay: time.Duration(cfg.RetryBaseDelaySeconds) * time.Second,
		Log:       log,
	}

	if chaos.DLQSnoop {
		rmq.RunDLQSnoop(context.Background(), log)
	}

	go func() {
		log.Info("rabbit_consumer_started", zap.String("queue", cfg.BrokerQueue))
		if err := rmq.Consume(context.Background(), overspend.HandleEnvelope); err != nil && err != context.Canceled {
			log.Error("consumer_stopped", zap.Error(err))
		}
	}()
	go jobWorker(context.Background(), overspend, postgres.NewJobRepo(store), cfg.OutboxBatchSize, cfg.OutboxMaxRetry, log)

	handler := httpapi.NewRouter(&httpapi.Server{
		Log:       log,
		JWTSecret: cfg.JWTSecret,
	})
	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
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

func jobWorker(ctx context.Context, svc *usecase.OverspendService, jobs repository.JobRepository, batch, maxRetry int, log *zap.Logger) {
	if batch <= 0 {
		batch = 20
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			items, err := jobs.FetchDue(ctx, batch)
			if err != nil {
				log.Error("fetch jobs", zap.Error(err))
				continue
			}
			for _, j := range items {
				if err := svc.ProcessJob(ctx, j); err != nil {
					retry := j.RetryCount + 1
					if retry >= maxRetry {
						_ = jobs.MarkFailed(ctx, j.ID)
						metrics.NotificationFailedTotal.Inc()
						log.Error("job_failed_max_retries",
							zap.String("job_id", j.ID.String()),
							zap.String("channel", j.Channel),
							zap.Int("retry_count", retry),
						)
						continue
					}
					metrics.NotificationRetryTotal.Inc()
					next := time.Now().UTC().Add(svc.Backoff(j.RetryCount))
					_ = jobs.MarkRetry(ctx, j.ID, retry, next)
					log.Warn("job_retry_scheduled",
						zap.String("job_id", j.ID.String()),
						zap.Int("retry_count", retry),
						zap.Time("next_retry_at", next),
						zap.String("channel", j.Channel),
					)
					continue
				}
				_ = jobs.MarkSent(ctx, j.ID)
				metrics.NotificationSentTotal.Inc()
				log.Info("job_sent", zap.String("job_id", j.ID.String()), zap.String("channel", j.Channel))
			}
		}
	}
}
