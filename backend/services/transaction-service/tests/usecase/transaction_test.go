package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/transaction-service/internal/domain"
	"github.com/moon-eye/velune/services/transaction-service/internal/repository"
	"github.com/moon-eye/velune/services/transaction-service/internal/usecase"
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

func (m *mockLedger) UpdateTransaction(ctx context.Context, userID uuid.UUID, next *domain.Transaction, prevVersion int64) error {
	return m.err
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

type testTxRepo struct {
	getFn func(context.Context, uuid.UUID, uuid.UUID) (*domain.Transaction, error)
}

func (t *testTxRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error) {
	if t.getFn != nil {
		return t.getFn(ctx, userID, id)
	}
	return nil, nil
}
func (t *testTxRepo) List(ctx context.Context, userID uuid.UUID, f repository.TransactionFilter, limit, offset int) ([]domain.Transaction, int64, error) {
	return []domain.Transaction{}, 0, nil
}
func (t *testTxRepo) SumIncomeExpenseInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (int64, int64, error) {
	return 0, 0, nil
}
func (t *testTxRepo) SumExpensesByCategoryInRange(ctx context.Context, userID uuid.UUID, from, to time.Time, currency string) (map[uuid.UUID]int64, error) {
	return map[uuid.UUID]int64{}, nil
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
	s := &usecase.TransactionService{
		Ledger:       &mockLedger{},
		Transactions: &noopTxRepo{},
		Accounts:     &noopAccountRepo{},
		Categories:   &mockCategoryRepo{},
		Logger:       zap.NewNop(),
	}
	uid := uuid.New()
	_, err := s.Create(context.Background(), uid, usecase.CreateTransactionInput{
		AccountID:   uuid.New(),
		AmountMinor: -100,
		Currency:    "USD",
		Type:        "expense",
		OccurredAt:  time.Now().UTC(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "amount")
}

func TestTransactionService_Create_TransferMissingCounterparty(t *testing.T) {
	s := &usecase.TransactionService{
		Ledger:       &mockLedger{},
		Transactions: &noopTxRepo{},
		Accounts:     &noopAccountRepo{},
		Categories:   &mockCategoryRepo{},
		Logger:       zap.NewNop(),
	}
	uid := uuid.New()
	_, err := s.Create(context.Background(), uid, usecase.CreateTransactionInput{
		AccountID:   uuid.New(),
		AmountMinor: 100,
		Currency:    "USD",
		Type:        "transfer",
		OccurredAt:  time.Now().UTC(),
	})
	require.Error(t, err)
}

func TestTransactionService_Get_NotFound(t *testing.T) {
	s := &usecase.TransactionService{
		Ledger:       &mockLedger{},
		Transactions: &testTxRepo{},
		Accounts:     &noopAccountRepo{},
		Categories:   &mockCategoryRepo{},
		Logger:       zap.NewNop(),
	}
	_, err := s.Get(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTransactionService_Update_VersionConflict(t *testing.T) {
	userID := uuid.New()
	txID := uuid.New()
	s := &usecase.TransactionService{
		Ledger: &mockLedger{},
		Transactions: &testTxRepo{getFn: func(ctx context.Context, u uuid.UUID, id uuid.UUID) (*domain.Transaction, error) {
			return &domain.Transaction{ID: txID, UserID: userID, Version: 2, AccountID: uuid.New(), AmountMinor: 100, Currency: "USD", Type: domain.TransactionExpense, OccurredAt: time.Now().UTC()}, nil
		}},
		Accounts:   &noopAccountRepo{},
		Categories: &mockCategoryRepo{},
		Logger:     zap.NewNop(),
	}
	_, err := s.Update(context.Background(), userID, txID, 1, usecase.UpdateTransactionInput{
		AccountID:   uuid.New(),
		AmountMinor: 100,
		Currency:    "USD",
		Type:        "expense",
		OccurredAt:  time.Now().UTC(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version conflict")
}

func TestTransactionService_Create_Success(t *testing.T) {
	s := &usecase.TransactionService{
		Ledger:       &mockLedger{},
		Transactions: &noopTxRepo{},
		Accounts:     &noopAccountRepo{},
		Categories:   &mockCategoryRepo{},
		Logger:       zap.NewNop(),
	}
	uid := uuid.New()
	_, err := s.Create(context.Background(), uid, usecase.CreateTransactionInput{
		AccountID:   uuid.New(),
		AmountMinor: 100,
		Currency:    "USD",
		Type:        "expense",
		OccurredAt:  time.Now().UTC(),
	})
	require.NoError(t, err)
}
