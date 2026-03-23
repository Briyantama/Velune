package httpapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
)

type budgetCreateReq struct {
	Name             string     `json:"name" validate:"required,min=1,max=200"`
	PeriodType       string     `json:"periodType" validate:"required,oneof=monthly weekly custom"`
	CategoryID       *uuid.UUID `json:"categoryId"`
	StartDate        time.Time  `json:"startDate" validate:"required"`
	EndDate          time.Time  `json:"endDate" validate:"required"`
	LimitAmountMinor int64      `json:"limitAmountMinor"`
	Currency         string     `json:"currency" validate:"required,len=3"`
}

type budgetUpdateReq struct {
	Name             string     `json:"name" validate:"required,min=1,max=200"`
	PeriodType       string     `json:"periodType" validate:"required,oneof=monthly weekly custom"`
	CategoryID       *uuid.UUID `json:"categoryId"`
	StartDate        time.Time  `json:"startDate" validate:"required"`
	EndDate          time.Time  `json:"endDate" validate:"required"`
	LimitAmountMinor int64      `json:"limitAmountMinor"`
	Currency         string     `json:"currency" validate:"required,len=3"`
}

func (s *Server) createBudget(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	var req budgetCreateReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	b, err := s.Budgets.Create(r.Context(), uid, usecase.CreateBudgetInput{
		Name:             req.Name,
		PeriodType:       req.PeriodType,
		CategoryID:       req.CategoryID,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		LimitAmountMinor: req.LimitAmountMinor,
		Currency:         req.Currency,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, b)
}

func (s *Server) listBudgets(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	page, limit := parsePageLimit(r)
	var activeOn *time.Time
	if v := r.URL.Query().Get("activeOn"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			WriteError(w, errs.New("VALIDATION_ERROR", "activeOn must be RFC3339", http.StatusBadRequest))
			return
		}
		activeOn = &t
	}
	list, total, err := s.Budgets.List(r.Context(), uid, page, limit, activeOn)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": list, "total": total, "page": page, "limit": limit})
}

func (s *Server) updateBudget(w http.ResponseWriter, r *http.Request) {
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
	var req budgetUpdateReq
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
	b, err := s.Budgets.Update(r.Context(), uid, id, v, usecase.UpdateBudgetInput{
		Name:             req.Name,
		PeriodType:       req.PeriodType,
		CategoryID:       req.CategoryID,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		LimitAmountMinor: req.LimitAmountMinor,
		Currency:         req.Currency,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

func (s *Server) deleteBudget(w http.ResponseWriter, r *http.Request) {
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
	if err := s.Budgets.Delete(r.Context(), uid, id, v); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
