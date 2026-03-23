package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/moon-eye/velune/services/auth-service/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
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
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if err := s.Validate.Struct(req); err != nil {
		writeErr(w, errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest))
		return
	}

	resp, err := s.Auth.Register(r.Context(), usecase.RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		BaseCurrency: req.BaseCurrency,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if err := s.Validate.Struct(req); err != nil {
		writeErr(w, errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest))
		return
	}

	resp, err := s.Auth.Login(r.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if err := s.Validate.Struct(req); err != nil {
		writeErr(w, errs.New("VALIDATION_ERROR", err.Error(), http.StatusBadRequest))
		return
	}

	resp, err := s.Auth.Refresh(r.Context(), usecase.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" || !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		writeErr(w, errs.ErrUnauthorized)
		return
	}
	token = strings.TrimSpace(token[len("bearer "):])
	if token == "" {
		writeErr(w, errs.ErrUnauthorized)
		return
	}

	resp, err := s.Auth.Me(r.Context(), token)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	ae, ok := err.(*errs.AppError)
	if ok && ae != nil {
		writeJSON(w, ae.Status, map[string]string{"code": ae.Code, "message": ae.Message})
		return
	}
	writeJSON(w, errs.ErrInternal.Status, map[string]string{
		"code":    errs.ErrInternal.Code,
		"message": errs.ErrInternal.Message,
	})
}
