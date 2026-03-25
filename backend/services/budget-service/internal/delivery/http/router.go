package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	txclient "github.com/moon-eye/velune/services/budget-service/internal/infrastructure/transactions"
	"github.com/moon-eye/velune/services/budget-service/internal/usecase"
	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
	"go.uber.org/zap"
)

type Server struct {
	Budgets          *usecase.BudgetService
	TxClient         *txclient.Client
	Validate         *validator.Validate
	Log              *zap.Logger
	JWTSecret        string
	AdminInternalKey string
	DB               *pgxpool.Pool
}

func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CorrelationIDHeader)

	r.Get("/health", s.health)
	r.Handle("/metrics", metrics.Handler())

	if s.AdminInternalKey != "" {
		r.Route("/internal/admin", func(r chi.Router) {
			r.Use(middlewares.AdminAPIKeyAuth(s.AdminInternalKey, s.Log))
			r.Post("/reconcile/budget", s.postAdminReconcileBudget)
		})
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.JWTAuth(s.JWTSecret, s.Log))
			r.Route("/budgets", func(r chi.Router) {
				r.Post("/", s.createBudget)
				r.Get("/", s.listBudgets)
				r.Put("/{id}", s.updateBudget)
				r.Delete("/{id}", s.deleteBudget)
				r.Get("/{id}/usage", s.getBudgetUsage)
			})
		})
	})
	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if s.DB != nil {
		if err := s.DB.Ping(r.Context()); err != nil {
			httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "down"})
			return
		}
	}
	httpx.WriteJSON(w, constx.StatusOK, map[string]string{"status": "ok"})
}
