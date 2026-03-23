package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
)

func scanTransaction(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Transaction, error) {
	var t domain.Transaction
	var typ string
	err := row.Scan(
		&t.ID, &t.UserID, &t.AccountID, &t.CategoryID, &t.CounterpartyAccountID, &t.AmountMinor, &t.Currency,
		&typ, &t.Description, &t.OccurredAt, &t.Version, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Type = domain.TransactionType(typ)
	return &t, nil
}
