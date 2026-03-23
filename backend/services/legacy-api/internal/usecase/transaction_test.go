package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockLedger struct {
	err error
}

func (m *mockLedger) CreateTransaction(ctx context.Context, t *domain.Transaction) error {
	return m.err
}

func (m *mockLedger) SoftDeleteTransaction(ctx context.Context, userID, id uuid.UUID, version int64) error {
	return nil
}

type mockCategoryRepo struct {
	cat *domain.Category
}

func (m *mockCategoryRepo) Create(ctx context.Context, c *domain.Category) error { return nil }
func (m *mockCategoryRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error) {
	return m.cat, nil
}
func (m *mockCategoryRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Category, int64, error) {
	return nil, 0, nil
}
func (m *mockCategoryRepo) Update(ctx context.Context, c *domain.Category) error { return nil }
func (m *mockCategoryRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	return nil
}

type noopTxRepo struct{}

func (n *noopTxRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error) {
	return nil, nil
}
func (n *noopTxRepo) List(ctx context.Context, userID uuid.UUID, f repository.TransactionFilter, limit, offset int) ([]domain.Transaction, int64, error) {
	return nil, 0, nil
}
func (n *noopTxRepo) SumIncomeExpenseInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (int64, int64, error) {
	return 0, 0, nil
}
func (n *noopTxRepo) SumExpensesByCategoryInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error) {
	return nil, nil
}

type noopAccountRepo struct{}

func (n *noopAccountRepo) Create(ctx context.Context, a *domain.Account) error { return nil }
func (n *noopAccountRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error) {
	return nil, nil
}
func (n *noopAccountRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Account, int64, error) {
	return nil, 0, nil
}
func (n *noopAccountRepo) Update(ctx context.Context, a *domain.Account) error { return nil }
func (n *noopAccountRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID, version int64) error {
	return nil
}

func TestTransactionService_Create_ExpenseNegativeAmount(t *testing.T) {
	s := &TransactionService{
		Ledger:       &mockLedger{},
		Transactions: &noopTxRepo{},
		Accounts:     &noopAccountRepo{},
		Categories:   &mockCategoryRepo{},
		Logger:       zap.NewNop(),
	}
	uid := uuid.New()
	_, err := s.Create(context.Background(), uid, CreateTransactionInput{
		AccountID:   uuid.New(),
		AmountMinor: -100,
		Currency:    "USD",
		Type:        "expense",
		OccurredAt:  time.Now(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "amount")
}

func TestTransactionService_Create_TransferMissingCounterparty(t *testing.T) {
	s := &TransactionService{
		Ledger:       &mockLedger{},
		Transactions: &noopTxRepo{},
		Accounts:     &noopAccountRepo{},
		Categories:   &mockCategoryRepo{},
		Logger:       zap.NewNop(),
	}
	uid := uuid.New()
	_, err := s.Create(context.Background(), uid, CreateTransactionInput{
		AccountID:   uuid.New(),
		AmountMinor: 100,
		Currency:    "USD",
		Type:        "transfer",
		OccurredAt:  time.Now(),
	})
	require.Error(t, err)
}
