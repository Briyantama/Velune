package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
	"go.uber.org/zap"
)

type Server struct {
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
			r.Get("/notifications/ping", ping)
		})
	})
	return r
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w,constx.StatusOK, map[string]string{"status": "ok"})
}

func ping(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w,constx.StatusOK, map[string]string{"status": "accepted"})
}
