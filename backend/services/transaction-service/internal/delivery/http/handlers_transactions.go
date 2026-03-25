package httpapi

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/usecase"
	"github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/contracts"
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

type transactionUpdateReq struct {
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
	httpx.WriteJSON(w, constx.StatusCreated, t)
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
		id, err := httpx.ParseUUID(v)
		if err != nil {
			httpx.WriteError(w, errs.New("VALIDATION_ERROR", "invalid accountId", constx.StatusBadRequest))
			return
		}
		accountID = &id
	}
	if v := q.Get("categoryId"); v != "" {
		id, err := httpx.ParseUUID(v)
		if err != nil {
			httpx.WriteError(w, errs.New("VALIDATION_ERROR", "invalid categoryId", constx.StatusBadRequest))
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
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{"items": list, "total": total, "page": page, "limit": limit})
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
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "version query is required", constx.StatusBadRequest))
		return
	}
	if err := s.Transactions.Delete(r.Context(), uid, id, v); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getTransaction(w http.ResponseWriter, r *http.Request) {
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
	t, err := s.Transactions.Get(r.Context(), uid, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, constx.StatusOK, t)
}

func (s *Server) updateTransaction(w http.ResponseWriter, r *http.Request) {
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
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "version query is required", constx.StatusBadRequest))
		return
	}
	var req transactionUpdateReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.ValidateStruct(&req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	t, err := s.Transactions.Update(r.Context(), uid, id, v, usecase.UpdateTransactionInput{
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
	httpx.WriteJSON(w, constx.StatusOK, t)
}

func (s *Server) transactionsSummary(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	from := httpx.ParseTimeQuery(r, "from")
	to := httpx.ParseTimeQuery(r, "to")
	if from == nil || to == nil {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "from and to are required RFC3339 values", constx.StatusBadRequest))
		return
	}
	cur := q.Get("currency")
	if cur == "" {
		cur = "USD"
	}
	inc, exp, err := s.Transactions.Summary(r.Context(), uid, *from, *to, cur)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, constx.StatusOK, contracts.TransactionSummary{
		From:         *from,
		To:           *to,
		Currency:     cur,
		IncomeMinor:  inc,
		ExpenseMinor: exp,
		NetMinor:     inc - exp,
	})
}

func (s *Server) transactionsSummaryByCategory(w http.ResponseWriter, r *http.Request) {
	uid, err := httpx.MustUserID(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	from := httpx.ParseTimeQuery(r, "from")
	to := httpx.ParseTimeQuery(r, "to")
	if from == nil || to == nil {
		httpx.WriteError(w, errs.New("VALIDATION_ERROR", "from and to are required RFC3339 values", constx.StatusBadRequest))
		return
	}
	cur := q.Get("currency")
	if cur == "" {
		cur = "USD"
	}
	m, err := s.Transactions.SummaryByCategory(r.Context(), uid, *from, *to, cur)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	breakdown := make([]contracts.TransactionCategorySummary, 0, len(m))
	for cid, total := range m {
		cidCopy := cid
		breakdown = append(breakdown, contracts.TransactionCategorySummary{
			CategoryID: &cidCopy,
			TotalMinor: total,
		})
	}
	httpx.WriteJSON(w, constx.StatusOK, contracts.TransactionCategoryTotalsResponse{
		From:      *from,
		To:        *to,
		Currency:  cur,
		Breakdown: breakdown,
	})
}
