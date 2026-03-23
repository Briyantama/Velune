package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/middlewares"
	"go.uber.org/zap"
)

type Server struct {
	Auth         *usecase.AuthService
	Accounts     *usecase.AccountService
	Categories   *usecase.CategoryService
	Transactions *usecase.TransactionService
	Budgets      *usecase.BudgetService
	Recurring    *usecase.RecurringService
	Reports      *usecase.ReportService
	Validate     *validator.Validate
	Log          *zap.Logger
	JWTSecret    string
	DB           *pgxpool.Pool
}

func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.RequestIDHeader)

	r.Get("/health", s.health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.JWTAuth(s.JWTSecret, s.Log))
			r.Route("/budgets", func(r chi.Router) {
				r.Post("/", s.createBudget)
				r.Get("/", s.listBudgets)
				r.Put("/{id}", s.updateBudget)
				r.Delete("/{id}", s.deleteBudget)
			})
		})
	})
	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if s.DB != nil {
		if err := s.DB.Ping(r.Context()); err != nil {
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "down"})
			return
		}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
