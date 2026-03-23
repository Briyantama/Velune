package httpapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
)

type transactionCreateReq struct {
	AccountID             uuid.UUID  `json:"accountId" validate:"required"`
	CategoryID            *uuid.UUID `json:"categoryId"`
	CounterpartyAccountID *uuid.UUID `json:"counterpartyAccountId"`
	AmountMinor           int64      `json:"amountMinor" validate:"required"`
	Currency              string     `json:"currency" validate:"required,len=3"`
	Type                  string     `json:"type" validate:"required,oneof=income expense transfer adjustment"`
	Description           string     `json:"description"`
	OccurredAt            time.Time  `json:"occurredAt" validate:"required"`
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var req transactionCreateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	t, err := s.Transactions.Create(r.Context(), uid, usecase.CreateTransactionInput{
		AccountID:             req.AccountID,
		CategoryID:            req.CategoryID,
		CounterpartyAccountID: req.CounterpartyAccountID,
		AmountMinor:           req.AmountMinor,
		Currency:              req.Currency,
		Type:                  req.Type,
		Description:           req.Description,
		OccurredAt:            req.OccurredAt,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, t)
}

func (s *Server) listTransactions(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	page, limit := httpx.ParsePageLimit(r)
	q := r.URL.Query()
	var accountID, categoryID *uuid.UUID
	if v := q.Get("accountId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpx.WriteError(w, errs.New("VALIDATION_ERROR", "invalid accountId", http.StatusBadRequest))
			return
		}
		accountID = &id
	}
	if v := q.Get("categoryId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpx.WriteError(w, errs.New("VALIDATION_ERROR", "invalid categoryId", http.StatusBadRequest))
			return
		}
		categoryID = &id
	}
	var txType *string
	if v := q.Get("type"); v != "" {
		txType = &v
	}
	in := usecase.ListTransactionsInput{
		Page:       page,
		Limit:      limit,
		AccountID:  accountID,
		CategoryID: categoryID,
		Type:       txType,
		From:       httpx.ParseTimeQuery(r, "from"),
		To:         httpx.ParseTimeQuery(r, "to"),
		Currency:   q.Get("currency"),
	}
	list, total, err := s.Transactions.List(r.Context(), uid, in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": list, "total": total, "page": page, "limit": limit})
}

func (s *Server) deleteTransaction(w http.ResponseWriter, r *http.Request) {
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
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "version query is required", http.StatusBadRequest))
		return
	}
	if err := s.Transactions.Delete(r.Context(), uid, id, v); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
