package httpapi

import (
	"net/http"

	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
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
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	tok, err := s.Auth.Register(r.Context(), usecase.RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		BaseCurrency: req.BaseCurrency,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, tok)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	tok, err := s.Auth.Login(r.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, tok)
}
