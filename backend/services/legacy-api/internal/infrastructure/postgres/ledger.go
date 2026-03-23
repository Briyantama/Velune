package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
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

	if err := l.insertTransaction(ctx, tx, t); err != nil {
		return err
	}
	if err := l.insertChangeEvent(ctx, tx, t.UserID, "transactions", t.ID, "CREATE", t.Version, t); err != nil {
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

	tr, err := l.loadTransactionForUpdate(ctx, tx, userID, id)
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

	const delQ = `
UPDATE transactions SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, delQ, id, userID, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOptimisticLock
	}
	return tx.Commit(ctx)
}

func (l *Ledger) loadTransactionForUpdate(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) (*domain.Transaction, error) {
	const q = `
SELECT id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type,
       description, occurred_at, version, created_at, updated_at, deleted_at
FROM transactions WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
FOR UPDATE`
	row := tx.QueryRow(ctx, q, id, userID)
	return scanTransaction(row)
}

func (l *Ledger) applyAccountDelta(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID, deltaMinor int64, currency string) error {
	const q = `
SELECT id, balance_minor, currency, version FROM accounts
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
FOR UPDATE`
	var id uuid.UUID
	var balance int64
	var cur string
	var ver int64
	err := tx.QueryRow(ctx, q, accountID, userID).Scan(&id, &balance, &cur, &ver)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}
	if err != nil {
		return err
	}
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
	const up = `
UPDATE accounts SET balance_minor = $1, version = version + 1, updated_at = now()
WHERE id = $2 AND user_id = $3 AND version = $4 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, up, newBal, accountID, userID, ver)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
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

func (l *Ledger) insertTransaction(ctx context.Context, tx pgx.Tx, t *domain.Transaction) error {
	const q = `
INSERT INTO transactions (id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type, description, occurred_at, version, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`
	_, err := tx.Exec(ctx, q,
		t.ID, t.UserID, t.AccountID, t.CategoryID, t.CounterpartyAccountID, t.AmountMinor, t.Currency, string(t.Type),
		t.Description, t.OccurredAt, t.Version, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (l *Ledger) insertChangeEvent(ctx context.Context, tx pgx.Tx, userID uuid.UUID, entityType string, entityID uuid.UUID, op string, ver int64, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	const q = `
INSERT INTO change_events (user_id, entity_type, entity_id, operation, version, payload)
VALUES ($1,$2,$3,$4,$5,$6)`
	_, err = tx.Exec(ctx, q, userID, entityType, entityID, op, ver, b)
	return err
}
