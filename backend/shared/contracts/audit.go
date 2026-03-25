package contracts

import "github.com/google/uuid"

const (
	EventBalanceMismatchDetected = "balance.mismatch.detected"
	EventBudgetMismatchDetected  = "budget.mismatch.detected"
)

// BalanceMismatchDetected is emitted when stored account balance disagrees with ledger sum.
type BalanceMismatchDetected struct {
	AccountID          uuid.UUID `json:"accountId"`
	UserID             uuid.UUID `json:"userId"`
	StoredBalanceMinor int64     `json:"storedBalanceMinor"`
	LedgerSumMinor     int64     `json:"ledgerSumMinor"`
	Currency           string    `json:"currency"`
}

// BudgetMismatchDetected is emitted when transaction-summary spend disagrees with alert-state implied spend.
type BudgetMismatchDetected struct {
	BudgetID                   uuid.UUID `json:"budgetId"`
	UserID                     uuid.UUID `json:"userId"`
	SpentFromTransactionsMinor int64     `json:"spentFromTransactionsMinor"`
	AlertImpliedSpentMinor     int64     `json:"alertImpliedSpentMinor"`
	LastUsagePercent           float64   `json:"lastUsagePercent"`
	Currency                   string    `json:"currency"`
}
