package httpapi

import (
	"net/http"

	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
)

type accountCreateReq struct {
	Name     string `json:"name" validate:"required,min=1,max=200"`
	Type     string `json:"type" validate:"required,oneof=wallet bank e_money cash card"`
	Currency string `json:"currency" validate:"required,len=3"`
}

type accountUpdateReq struct {
	Name string `json:"name" validate:"required,min=1,max=200"`
	Type string `json:"type" validate:"required,oneof=wallet bank e_money cash card"`
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	var req accountCreateReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	a, err := s.Accounts.Create(r.Context(), uid, usecase.CreateAccountInput{
		Name:     req.Name,
		Type:     req.Type,
		Currency: req.Currency,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, a)
}

func (s *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	page, limit := parsePageLimit(r)
	list, total, err := s.Accounts.List(r.Context(), uid, page, limit)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": list, "total": total, "page": page, "limit": limit})
}

func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}
	a, err := s.Accounts.Get(r.Context(), uid, id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, a)
}

func (s *Server) updateAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}
	var req accountUpdateReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	v, ok := parseInt64Query(r, "version")
	if !ok {
		WriteError(w, errs.New("VALIDATION_ERROR", "version query is required", http.StatusBadRequest))
		return
	}
	a, err := s.Accounts.Update(r.Context(), uid, id, v, usecase.UpdateAccountInput{
		Name: req.Name,
		Type: req.Type,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, a)
}

func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}
	v, ok := parseInt64Query(r, "version")
	if !ok {
		WriteError(w, errs.New("VALIDATION_ERROR", "version query is required", http.StatusBadRequest))
		return
	}
	if err := s.Accounts.Delete(r.Context(), uid, id, v); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
