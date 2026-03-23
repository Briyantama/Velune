package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moon-eye/velune/services/transaction-service/internal/usecase"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/middlewares"
	"go.uber.org/zap"
)

type Server struct {
	Accounts     *usecase.AccountService
	Categories   *usecase.CategoryService
	Transactions *usecase.TransactionService
	Recurring    *usecase.RecurringService
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
			r.Route("/accounts", func(r chi.Router) {
				r.Post("/", s.createAccount)
				r.Get("/", s.listAccounts)
				r.Get("/{id}", s.getAccount)
				r.Put("/{id}", s.updateAccount)
				r.Delete("/{id}", s.deleteAccount)
			})
			r.Route("/categories", func(r chi.Router) {
				r.Post("/", s.createCategory)
				r.Get("/", s.listCategories)
				r.Put("/{id}", s.updateCategory)
				r.Delete("/{id}", s.deleteCategory)
			})
			r.Route("/transactions", func(r chi.Router) {
				r.Post("/", s.createTransaction)
				r.Get("/", s.listTransactions)
				r.Get("/{id}", s.getTransaction)
				r.Patch("/{id}", s.updateTransaction)
				r.Delete("/{id}", s.deleteTransaction)
			})
			r.Route("/recurring", func(r chi.Router) {
				r.Post("/", s.createRecurring)
				r.Get("/", s.listRecurring)
				r.Delete("/{id}", s.deleteRecurring)
			})
			r.Get("/transactions/summary", s.transactionsSummary)
			r.Get("/transactions/summary/categories", s.transactionsSummaryByCategory)
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
