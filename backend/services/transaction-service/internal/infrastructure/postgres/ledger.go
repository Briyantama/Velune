package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
	"github.com/moon-eye/velune/services/transaction-service/internal/repository"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

type Ledger struct{ s *Store }

func NewLedger(s *Store) repository.Ledger {
	return &Ledger{s: s}
}

func (l *Ledger) CreateTransaction(ctx context.Context, t *domain.Transaction) error {
	tx, err := l.s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := l.s.Queries.WithTx(tx)

	switch t.Type {
	case domain.TransactionIncome:
		if err := l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, t.AmountMinor, t.Currency); err != nil {
			return err
		}
	case domain.TransactionExpense:
		if err := l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, -t.AmountMinor, t.Currency); err != nil {
			return err
		}
	case domain.TransactionTransfer:
		if t.CounterpartyAccountID == nil {
			return errors.New("counterparty account required for transfer")
		}
		if err := l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, -t.AmountMinor, t.Currency); err != nil {
			return err
		}
		if err := l.applyAccountDelta(ctx, tx, t.UserID, *t.CounterpartyAccountID, t.AmountMinor, t.Currency); err != nil {
			return err
		}
	case domain.TransactionAdjustment:
		if err := l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, t.AmountMinor, t.Currency); err != nil {
			return err
		}
	default:
		return errors.New("unsupported transaction type")
	}

	if err := l.insertTransaction(ctx, qtx, t); err != nil {
		return err
	}
	if err := l.insertLedgerEntries(ctx, qtx, t, "create"); err != nil {
		return err
	}
	if err := l.insertChangeEvent(ctx, qtx, t.UserID, "transactions", t.ID, "CREATE", t.Version, t); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (l *Ledger) UpdateTransaction(ctx context.Context, userID uuid.UUID, next *domain.Transaction, prevVersion int64) error {
	tx, err := l.s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := l.s.Queries.WithTx(tx)

	prev, err := l.loadTransactionForUpdate(ctx, qtx, userID, next.ID)
	if err != nil {
		return err
	}
	if prev == nil {
		return repository.ErrNotFound
	}
	if prev.Version != prevVersion {
		return repository.ErrOptimisticLock
	}

	// Reverse previous impact first.
	if err := l.applyReverseImpact(ctx, tx, prev); err != nil {
		return err
	}
	// Apply next impact.
	if err := l.applyForwardImpact(ctx, tx, next); err != nil {
		return err
	}

	tag, err := qtx.LedgerUpdateTransaction(ctx, db.LedgerUpdateTransactionParams{
		AccountID:             helper.ToPgUUID(next.AccountID),
		CategoryID:            helper.ToPgUUIDPtr(next.CategoryID),
		CounterpartyAccountID: helper.ToPgUUIDPtr(next.CounterpartyAccountID),
		AmountMinor:           next.AmountMinor,
		Currency:              next.Currency,
		Type:                  string(next.Type),
		Description:           next.Description,
		OccurredAt:            helper.ToPgTS(next.OccurredAt),
		Version:               next.Version,
		UpdatedAt:             helper.ToPgTS(next.UpdatedAt),
		ID:                    helper.ToPgUUID(next.ID),
		UserID:                helper.ToPgUUID(userID),
		Version_2:             prevVersion,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}

	if err := l.insertLedgerEntries(ctx, qtx, next, "update"); err != nil {
		return err
	}
	if err := l.insertChangeEvent(ctx, qtx, userID, "transactions", next.ID, "UPDATE", next.Version, map[string]any{
		"before": prev,
		"after":  next,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (l *Ledger) SoftDeleteTransaction(ctx context.Context, userID, id uuid.UUID, version int64) error {
	tx, err := l.s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := l.s.Queries.WithTx(tx)

	tr, err := l.loadTransactionForUpdate(ctx, qtx, userID, id)
	if err != nil {
		return err
	}
	if tr == nil {
		return repository.ErrNotFound
	}
	if tr.Version != version {
		return repository.ErrOptimisticLock
	}

	switch tr.Type {
	case domain.TransactionIncome:
		if err := l.applyAccountDelta(ctx, tx, userID, tr.AccountID, -tr.AmountMinor, tr.Currency); err != nil {
			return err
		}
	case domain.TransactionExpense:
		if err := l.applyAccountDelta(ctx, tx, userID, tr.AccountID, tr.AmountMinor, tr.Currency); err != nil {
			return err
		}
	case domain.TransactionTransfer:
		if tr.CounterpartyAccountID == nil {
			return errors.New("invalid transfer row")
		}
		if err := l.applyAccountDelta(ctx, tx, userID, tr.AccountID, tr.AmountMinor, tr.Currency); err != nil {
			return err
		}
		if err := l.applyAccountDelta(ctx, tx, userID, *tr.CounterpartyAccountID, -tr.AmountMinor, tr.Currency); err != nil {
			return err
		}
	case domain.TransactionAdjustment:
		if err := l.applyAccountDelta(ctx, tx, userID, tr.AccountID, -tr.AmountMinor, tr.Currency); err != nil {
			return err
		}
	}
	if err := l.insertLedgerEntries(ctx, qtx, tr, "delete"); err != nil {
		return err
	}

	tag, err := qtx.LedgerSoftDeleteTransaction(ctx, db.LedgerSoftDeleteTransactionParams{
		ID:      helper.ToPgUUID(id),
		UserID:  helper.ToPgUUID(userID),
		Version: version,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}
	if err := l.insertChangeEvent(ctx, qtx, userID, "transactions", id, "DELETE", version+1, tr); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (l *Ledger) loadTransactionForUpdate(ctx context.Context, qtx *db.Queries, userID, id uuid.UUID) (*domain.Transaction, error) {
	row, err := qtx.LedgerLoadTransactionForUpdate(ctx, db.LedgerLoadTransactionForUpdateParams{
		ID:     helper.ToPgUUID(id),
		UserID: helper.ToPgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return transactionFromGetByIDRow(db.TransactionGetByIDRow{
		ID:                    row.ID,
		UserID:                row.UserID,
		AccountID:             row.AccountID,
		CategoryID:            row.CategoryID,
		CounterpartyAccountID: row.CounterpartyAccountID,
		AmountMinor:           row.AmountMinor,
		Currency:              row.Currency,
		Type:                  row.Type,
		Description:           row.Description,
		OccurredAt:            row.OccurredAt,
		Version:               row.Version,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		DeletedAt:             row.DeletedAt,
	}), nil
}

func (l *Ledger) applyAccountDelta(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID, deltaMinor int64, currency string) error {
	row, err := l.s.Queries.WithTx(tx).LedgerSelectAccountForUpdate(ctx, db.LedgerSelectAccountForUpdateParams{
		ID:     helper.ToPgUUID(accountID),
		UserID: helper.ToPgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}
	if err != nil {
		return err
	}
	balance := row.BalanceMinor
	cur := row.Currency
	ver := row.Version
	if cur != currency {
		return errors.New("currency mismatch")
	}
	newBal, err2 := addInt64(balance, deltaMinor)
	if err2 != nil {
		return err2
	}
	if newBal < 0 {
		return repository.ErrInsufficientBalance
	}
	tag, err := l.s.Queries.WithTx(tx).LedgerUpdateAccountBalanceVersion(ctx, db.LedgerUpdateAccountBalanceVersionParams{
		BalanceMinor: newBal,
		ID:           helper.ToPgUUID(accountID),
		UserID:       helper.ToPgUUID(userID),
		Version:      ver,
	})
	if err != nil {
		return err
	}
	if tag == 0 {
		return repository.ErrOptimisticLock
	}
	return nil
}

func addInt64(a, b int64) (int64, error) {
	c := a + b
	if (b > 0 && c < a) || (b < 0 && c > a) {
		return 0, errors.New("integer overflow")
	}
	return c, nil
}

func (l *Ledger) insertTransaction(ctx context.Context, qtx *db.Queries, t *domain.Transaction) error {
	return qtx.LedgerInsertTransaction(ctx, db.LedgerInsertTransactionParams{
		ID:                    helper.ToPgUUID(t.ID),
		UserID:                helper.ToPgUUID(t.UserID),
		AccountID:             helper.ToPgUUID(t.AccountID),
		CategoryID:            helper.ToPgUUIDPtr(t.CategoryID),
		CounterpartyAccountID: helper.ToPgUUIDPtr(t.CounterpartyAccountID),
		AmountMinor:           t.AmountMinor,
		Currency:              t.Currency,
		Type:                  string(t.Type),
		Description:           t.Description,
		OccurredAt:            helper.ToPgTS(t.OccurredAt),
		Version:               t.Version,
		CreatedAt:             helper.ToPgTS(t.CreatedAt),
		UpdatedAt:             helper.ToPgTS(t.UpdatedAt),
	})
}

func (l *Ledger) insertChangeEvent(ctx context.Context, qtx *db.Queries, userID uuid.UUID, entityType string, entityID uuid.UUID, op string, ver int64, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return qtx.LedgerInsertChangeEvent(ctx, db.LedgerInsertChangeEventParams{
		UserID:     helper.ToPgUUID(userID),
		EntityType: entityType,
		EntityID:   helper.ToPgUUID(entityID),
		Operation:  op,
		Version:    ver,
		Payload:    b,
	})
}

func (l *Ledger) insertLedgerEntries(ctx context.Context, qtx *db.Queries, t *domain.Transaction, reason string) error {
	insert := func(accountID uuid.UUID, direction string, amount int64) error {
		return qtx.LedgerInsertLedgerEntry(ctx, db.LedgerInsertLedgerEntryParams{
			TransactionID: helper.ToPgUUID(t.ID),
			UserID:        helper.ToPgUUID(t.UserID),
			AccountID:     helper.ToPgUUID(accountID),
			Direction:     direction,
			AmountMinor:   amount,
			Currency:      t.Currency,
			Reason:        reason,
		})
	}
	switch t.Type {
	case domain.TransactionIncome:
		return insert(t.AccountID, "credit", t.AmountMinor)
	case domain.TransactionExpense:
		return insert(t.AccountID, "debit", t.AmountMinor)
	case domain.TransactionTransfer:
		if t.CounterpartyAccountID == nil {
			return errors.New("counterparty account required for transfer")
		}
		if err := insert(t.AccountID, "debit", t.AmountMinor); err != nil {
			return err
		}
		return insert(*t.CounterpartyAccountID, "credit", t.AmountMinor)
	case domain.TransactionAdjustment:
		dir := "credit"
		amt := t.AmountMinor
		if amt < 0 {
			dir = "debit"
			amt = -amt
		}
		return insert(t.AccountID, dir, amt)
	default:
		return errors.New("unsupported transaction type")
	}
}

func (l *Ledger) applyForwardImpact(ctx context.Context, tx pgx.Tx, t *domain.Transaction) error {
	switch t.Type {
	case domain.TransactionIncome:
		return l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, t.AmountMinor, t.Currency)
	case domain.TransactionExpense:
		return l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, -t.AmountMinor, t.Currency)
	case domain.TransactionTransfer:
		if t.CounterpartyAccountID == nil {
			return errors.New("counterparty account required for transfer")
		}
		if err := l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, -t.AmountMinor, t.Currency); err != nil {
			return err
		}
		return l.applyAccountDelta(ctx, tx, t.UserID, *t.CounterpartyAccountID, t.AmountMinor, t.Currency)
	case domain.TransactionAdjustment:
		return l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, t.AmountMinor, t.Currency)
	default:
		return errors.New("unsupported transaction type")
	}
}

func (l *Ledger) applyReverseImpact(ctx context.Context, tx pgx.Tx, t *domain.Transaction) error {
	switch t.Type {
	case domain.TransactionIncome:
		return l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, -t.AmountMinor, t.Currency)
	case domain.TransactionExpense:
		return l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, t.AmountMinor, t.Currency)
	case domain.TransactionTransfer:
		if t.CounterpartyAccountID == nil {
			return errors.New("invalid transfer row")
		}
		if err := l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, t.AmountMinor, t.Currency); err != nil {
			return err
		}
		return l.applyAccountDelta(ctx, tx, t.UserID, *t.CounterpartyAccountID, -t.AmountMinor, t.Currency)
	case domain.TransactionAdjustment:
		return l.applyAccountDelta(ctx, tx, t.UserID, t.AccountID, -t.AmountMinor, t.Currency)
	default:
		return errors.New("unsupported transaction type")
	}
}
