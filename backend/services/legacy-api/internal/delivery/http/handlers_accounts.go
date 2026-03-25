package httpapi

import (
	"net/http"

	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
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
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req accountCreateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	a, err := s.Accounts.Create(r.Context(), uid, usecase.CreateAccountInput{
		Name:     req.Name,
		Type:     req.Type,
		Currency: req.Currency,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w,constx.StatusCreated, a)
}

func (s *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	page, limit := httpx.ParsePageLimit(r)
	list, total, err := s.Accounts.List(r.Context(), uid, page, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w,constx.StatusOK, map[string]any{"items": list, "total": total, "page": page, "limit": limit})
}

func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := httpx.ParseUUID(r.URL.Query().Get("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	a, err := s.Accounts.Get(r.Context(), uid, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w,constx.StatusOK, a)
}

func (s *Server) updateAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := httpx.ParseUUID(r.URL.Query().Get("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req accountUpdateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	v, ok := httpx.ParseInt64Query(r, "version")
	if !ok {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "version query is required",constx.StatusBadRequest))
		return
	}
	a, err := s.Accounts.Update(r.Context(), uid, id, v, usecase.UpdateAccountInput{
		Name: req.Name,
		Type: req.Type,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w,constx.StatusOK, a)
}

func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := httpx.ParseUUID(r.URL.Query().Get("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	v, ok := httpx.ParseInt64Query(r, "version")
	if !ok {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "version query is required",constx.StatusBadRequest))
		return
	}
	if err := s.Accounts.Delete(r.Context(), uid, id, v); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(constx.StatusNoContent)
}
