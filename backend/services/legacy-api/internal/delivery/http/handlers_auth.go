package httpapi

import (
	"net/http"

	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/httpx"
)

type registerReq struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8,max=72"`
	BaseCurrency string `json:"baseCurrency" validate:"required,len=3"`
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	tok, err := s.Auth.Register(r.Context(), usecase.RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		BaseCurrency: req.BaseCurrency,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w,constx.StatusCreated, tok)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	tok, err := s.Auth.Login(r.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w,constx.StatusOK, tok)
}
