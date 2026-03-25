package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	httpapi "github.com/moon-eye/velune/services/auth-service/internal/delivery/http"
	postgres "github.com/moon-eye/velune/services/auth-service/internal/infrastructure/postgres"
	"github.com/moon-eye/velune/services/auth-service/internal/usecase"
	sharedconfig "github.com/moon-eye/velune/shared/config"
	"github.com/moon-eye/velune/shared/helper"
	sharedlog "github.com/moon-eye/velune/shared/logger"
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

	authSvc := &usecase.AuthService{
		Users:         userRepo,
		RefreshTokens: refreshRepo,
		JWTSecret:     cfg.JWTSecret,
		AccessTTL:     cfg.JWTExpiry,
		// Refresh token TTL is configurable via REFRESH_TOKEN_TTL (see usecase).
	}

	v := validator.New()
	server := &httpapi.Server{Auth: authSvc, Validate: v}
	handler := httpapi.NewRouter(server)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Get("/health", health())
	r.Mount("/", handler)

	addr := cfg.HTTPHost + ":" + cfg.HTTPPort
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           r,
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
		w.Header().Set("Content-Type", "application/json")
		_ = helper.EncodeJSON(w, map[string]any{
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
