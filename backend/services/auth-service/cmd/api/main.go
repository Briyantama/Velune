package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	httpapi "github.com/moon-eye/velune/services/auth-service/internal/delivery/http"
	email "github.com/moon-eye/velune/services/auth-service/internal/infrastructure/email"
	postgres "github.com/moon-eye/velune/services/auth-service/internal/infrastructure/postgres"
	"github.com/moon-eye/velune/services/auth-service/internal/usecase"
	sharedconfig "github.com/moon-eye/velune/shared/config"
	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/events"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
	"github.com/moon-eye/velune/shared/otelx"
	"github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/sim"
	stringx "github.com/moon-eye/velune/shared/stringx"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func main() {
	log, err := sharedlog.New("auth-service")
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	defer log.Sync()

	cfg, err := sharedconfig.Load("auth-service")
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

	userRepo := postgres.NewUserRepo(store)
	refreshRepo := postgres.NewRefreshTokenRepo(store)

	otpRepo := postgres.NewOTPVerificationRepo(store)
	provisioningRepo := postgres.NewProvisioningStateRepo(store)

	otpValiditySec := 300
	if v := os.Getenv("OTP_VALIDITY_SECONDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			otpValiditySec = i
		}
	}
	otpResendCooldownSec := 30
	if v := os.Getenv("OTP_RESEND_COOLDOWN_SECONDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			otpResendCooldownSec = i
		}
	}
	otpMaxResends := 3
	if v := os.Getenv("OTP_MAX_RESENDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			otpMaxResends = i
		}
	}
	otpMaxVerifyAttempts := 3
	if v := os.Getenv("OTP_MAX_VERIFY_ATTEMPTS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 1 {
			otpMaxVerifyAttempts = i
		}
	}

	// Email delivery: use SMTP if configured; otherwise fall back to a stub in dev.
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")
	if smtpFrom == "" {
		smtpFrom = cfg.EmailFrom
	}
	smtpTLS := stringx.StringsEqualTrue(os.Getenv("SMTP_TLS"))

	var otpSender usecase.OTPSender
	if smtpHost == "" || smtpPort == "" || smtpFrom == "" {
		if cfg.Environment == "production" {
			log.Fatal("smtp not configured: set SMTP_HOST/SMTP_PORT/SMTP_FROM")
		}
		otpSender = &email.StubOtpSender{Log: log, From: smtpFrom}
	} else {
		otpSender = email.NewSMTPOtpSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom, smtpTLS, log)
	}

	// RabbitMQ publishing + provisioning completion consumer.
	var publisherPtr *events.RabbitPublisher
	chaosCfg := sim.LoadFromEnv()
	pub, err := events.NewRabbitPublisher(
		cfg.BrokerURL,
		cfg.BrokerExchange,
		cfg.BrokerRoutingKey,
		cfg.BrokerDLX,
		cfg.BrokerDLQRoutingKey,
		chaosCfg,
	)
	if err != nil {
		log.Error("rabbit_publisher_init_failed", zap.Error(err))
	} else {
		publisherPtr = pub
		defer func() { _ = publisherPtr.Close() }()

		log.Info("rabbit_publisher_ready", zap.String("exchange", cfg.BrokerExchange))
	}

	queueName := os.Getenv("AUTH_PROVISION_COMPLETED_QUEUE")
	if queueName == "" {
		queueName = "auth.provision_completed"
	}
	completionRoutingKey := contracts.EventUserFirstLoginProvisionCompleted
	exchange := cfg.BrokerExchange
	dlx := cfg.BrokerDLX
	dlqRoutingKey := cfg.BrokerDLQRoutingKey

	// Completion consumer only updates auth-service provisioning state (idempotent upsert).
	go func() {
		if cfg.BrokerURL == "" || exchange == "" {
			return
		}
		conn, err := amqp.Dial(cfg.BrokerURL)
		if err != nil {
			log.Error("rabbit_consumer_dial_failed", zap.Error(err))
			return
		}
		defer func() { _ = conn.Close() }()
		ch, err := conn.Channel()
		if err != nil {
			log.Error("rabbit_consumer_channel_failed", zap.Error(err))
			return
		}
		defer func() { _ = ch.Close() }()

		if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
			log.Error("rabbit_consumer_exchange_declare_failed", zap.Error(err))
			return
		}

		args := amqp.Table{}
		if dlx != "" {
			args["x-dead-letter-exchange"] = dlx
			if dlqRoutingKey != "" {
				args["x-dead-letter-routing-key"] = dlqRoutingKey
			}
		}
		q, err := ch.QueueDeclare(queueName, true, false, false, false, args)
		if err != nil {
			log.Error("rabbit_consumer_queue_declare_failed", zap.Error(err))
			return
		}
		if err := ch.QueueBind(q.Name, completionRoutingKey, exchange, false, nil); err != nil {
			log.Error("rabbit_consumer_queue_bind_failed", zap.Error(err))
			return
		}

		msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
		if err != nil {
			log.Error("rabbit_consumer_consume_failed", zap.Error(err))
			return
		}

		log.Info("rabbit_completion_consumer_started", zap.String("queue", q.Name), zap.String("routingKey", completionRoutingKey))
		for msg := range msgs {
			var env contracts.EventEnvelope
			if err := json.Unmarshal(msg.Body, &env); err != nil {
				_ = msg.Nack(false, true)
				continue
			}
			if env.EventType != contracts.EventUserFirstLoginProvisionCompleted {
				_ = msg.Ack(false)
				continue
			}

			var payload contracts.UserFirstLoginProvisionCompleted
			if err := json.Unmarshal(env.Payload, &payload); err != nil {
				_ = msg.Nack(false, true)
				continue
			}

			if err := provisioningRepo.MarkAccountProvisionedAt(context.Background(), payload.UserID, payload.OccurredAt); err != nil {
				log.Error("provisioning_state_mark_failed", zap.Error(err), zap.String("user_id", payload.UserID.String()))
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}()

	authSvc := &usecase.AuthService{
		Log:                  log,
		Users:                userRepo,
		RefreshTokens:        refreshRepo,
		OTPVerifications:     otpRepo,
		ProvisioningState:    provisioningRepo,
		OTPSender:            otpSender,
		JWTSecret:            cfg.JWTSecret,
		AccessTTL:            cfg.JWTExpiry,
		OTPValidity:          time.Duration(otpValiditySec) * time.Second,
		OTPResendCooldown:    time.Duration(otpResendCooldownSec) * time.Second,
		OTPMaxResends:        otpMaxResends,
		OTPMaxVerifyAttempts: otpMaxVerifyAttempts,
		EventPublisher:       publisherPtr,
		// Refresh token TTL is configurable via REFRESH_TOKEN_TTL (see usecase).
	}

	v := validator.New()
	server := &httpapi.Server{Auth: authSvc, Validate: v}
	handler := httpapi.NewRouter(server)

	r := chi.NewRouter()
	r.Use(middlewares.CorrelationIDHeader)
	r.Use(middleware.RealIP, middleware.Recoverer)
	r.Get("/health", health())
	r.Handle("/metrics", metrics.Handler())
	r.Mount("/", handler)

	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           otelx.HTTPHandler(r, "http.server"),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("listening", zap.String("addr", addr))
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

func health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, constx.StatusOK, map[string]string{
			"status":  "ok",
			"service": "auth-service",
		})
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
