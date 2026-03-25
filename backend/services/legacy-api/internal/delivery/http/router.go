package httpapi

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
)

// NewRouter serves health, Prometheus metrics, and explicit 410 responses for migrated /api/v1 paths.
func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CorrelationIDHeader)

	r.Get("/health", health)
	r.Handle("/metrics", metrics.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.NotFound(apiMigratedGone)
	})
	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{
		"status":  "ok",
		"service": "legacy-api",
		"role":    "strangler_shell",
	})
}

func apiMigratedGone(w http.ResponseWriter, _ *http.Request) {
	// 410 Gone — HTTP semantic for retired resources
	httpx.WriteJSON(w, http.StatusGone, map[string]any{
		"code":    "ENDPOINT_MOVED",
		"message": "Business APIs are served via api-gateway to microservices; legacy surface retired.",
		"docs":    os.Getenv("VELOUNE_MIGRATION_DOCS_URL"),
	})
}
