package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/moon-eye/velune/services/auth-service/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/stringx"
)

type Server struct {
	Auth     *usecase.AuthService
	Validate *validator.Validate
}

func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/v1/auth/register", s.handleRegister)
	r.Post("/api/v1/auth/login", s.handleLogin)
	r.Post("/api/v1/auth/refresh", s.handleRefresh)
	r.Get("/api/v1/auth/me", s.handleMe)
	return r
}

type registerReq struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8,max=72"`
	BaseCurrency string `json:"baseCurrency" validate:"required,len=3"`
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := s.Validate.Struct(req); err != nil {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest))
		return
	}

	resp, err := s.Auth.Register(r.Context(), usecase.RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		BaseCurrency: req.BaseCurrency,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := s.Validate.Struct(req); err != nil {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest))
		return
	}

	resp, err := s.Auth.Login(r.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := s.Validate.Struct(req); err != nil {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest))
		return
	}

	resp, err := s.Auth.Refresh(r.Context(), usecase.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	token := stringx.TrimSpace(r.Header.Get("Authorization"))
	if token == "" || !stringx.HasPrefix(stringx.Lower(token), "bearer ") {
		httpx.WriteError(w, errs.ErrUnauthorized)
		return
	}
	token = stringx.TrimSpace(token[len("bearer "):])
	if token == "" {
		httpx.WriteError(w, errs.ErrUnauthorized)
		return
	}

	resp, err := s.Auth.Me(r.Context(), token)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}
