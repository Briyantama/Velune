package httpapi

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/moon-eye/velune/services/report-service/internal/usecase"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
	"go.uber.org/zap"
)

type Server struct {
	Reports   *usecase.ReportService
	Validate  *validator.Validate
	Log       *zap.Logger
	JWTSecret string
}

func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CorrelationIDHeader)
	r.Get("/health", health)
	r.Handle("/metrics", metrics.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.JWTAuth(s.JWTSecret, s.Log))
			r.Get("/reports/monthly", s.monthlyReport)
		})
	})
	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, constx.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) monthlyReport(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	y, err := strconv.Atoi(q.Get("year"))
	if err != nil || y < 1900 || y > 3000 {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "valid year is required", constx.StatusBadRequest))
		return
	}
	m, err := strconv.Atoi(q.Get("month"))
	if err != nil || m < 1 || m > 12 {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "valid month 1-12 is required", constx.StatusBadRequest))
		return
	}
	rep, err := s.Reports.Monthly(r.Context(), uid, usecase.MonthlyInput{
		Year:     y,
		Month:    m,
		Currency: q.Get("currency"),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, constx.StatusOK, rep)
}
