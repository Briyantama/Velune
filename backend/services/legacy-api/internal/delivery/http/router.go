package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
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
	r.Use(RequestIDHeader)

	r.Get("/health", s.health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(JWTAuth(s.JWTSecret, s.Log))
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
				r.Delete("/{id}", s.deleteTransaction)
			})
			r.Route("/budgets", func(r chi.Router) {
				r.Post("/", s.createBudget)
				r.Get("/", s.listBudgets)
				r.Put("/{id}", s.updateBudget)
				r.Delete("/{id}", s.deleteBudget)
			})
			r.Route("/recurring", func(r chi.Router) {
				r.Post("/", s.createRecurring)
				r.Get("/", s.listRecurring)
				r.Delete("/{id}", s.deleteRecurring)
			})
			r.Get("/reports/monthly", s.monthlyReport)
		})
	})
	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if s.DB != nil {
		if err := s.DB.Ping(r.Context()); err != nil {
			WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "down"})
			return
		}
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parsePageLimit(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	return page, limit
}

func mustUserID(r *http.Request) (uuid.UUID, error) {
	uid, ok := UserID(r.Context())
	if !ok || uid == uuid.Nil {
		return uuid.Nil, errs.ErrUnauthorized
	}
	return uid, nil
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
	}
	return nil
}

func validateStruct(s *Server, v any) error {
	if err := s.Validate.Struct(v); err != nil {
		return errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
	}
	return nil
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errs.New("VALIDATION_ERROR", "invalid id", http.StatusBadRequest)
	}
	return id, nil
}

func parseTimeQuery(r *http.Request, key string) *time.Time {
	if v := r.URL.Query().Get(key); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
}

func parseInt64Query(r *http.Request, key string) (int64, bool) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
