package httpapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
)

type recurringCreateReq struct {
	AccountID   uuid.UUID  `json:"accountId" validate:"required"`
	CategoryID  *uuid.UUID `json:"categoryId"`
	AmountMinor int64      `json:"amountMinor" validate:"min=1"`
	Currency    string     `json:"currency" validate:"required,len=3"`
	Type        string     `json:"type" validate:"required,oneof=income expense"`
	Frequency   string     `json:"frequency" validate:"required,oneof=daily weekly monthly yearly"`
	NextRunAt   time.Time  `json:"nextRunAt" validate:"required"`
	Description string     `json:"description"`
}

func (s *Server) createRecurring(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	var req recurringCreateReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	rr, err := s.Recurring.Create(r.Context(), uid, usecase.CreateRecurringInput{
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		AmountMinor: req.AmountMinor,
		Currency:    req.Currency,
		Type:        req.Type,
		Frequency:   req.Frequency,
		NextRunAt:   req.NextRunAt,
		Description: req.Description,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, rr)
}

func (s *Server) listRecurring(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	page, limit := parsePageLimit(r)
	activeOnly := r.URL.Query().Get("activeOnly") == "true"
	list, total, err := s.Recurring.List(r.Context(), uid, page, limit, activeOnly)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": list, "total": total, "page": page, "limit": limit})
}

func (s *Server) deleteRecurring(w http.ResponseWriter, r *http.Request) {
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
	if err := s.Recurring.Delete(r.Context(), uid, id, v); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
