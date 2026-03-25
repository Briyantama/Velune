package contracts

import (
	"time"

	"github.com/google/uuid"
)

type TransactionAnalyticsQuery struct {
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Currency string    `json:"currency"`
}

type TransactionCategorySummary struct {
	CategoryID *uuid.UUID `json:"categoryId,omitempty"`
	TotalMinor int64      `json:"totalMinor"`
}

type TransactionCategoryTotalsResponse struct {
	From      time.Time                    `json:"from"`
	To        time.Time                    `json:"to"`
	Currency  string                       `json:"currency"`
	Breakdown []TransactionCategorySummary `json:"breakdown"`
}

type MonthlyCategoryBreakdownDTO struct {
	CategoryID   *uuid.UUID `json:"categoryId,omitempty"`
	CategoryName string     `json:"categoryName"`
	TotalMinor   int64      `json:"totalMinor"`
	Currency     string     `json:"currency"`
}

type MonthlyReportDTO struct {
	UserID       uuid.UUID                     `json:"userId"`
	Year         int                           `json:"year"`
	Month        int                           `json:"month"`
	IncomeMinor  int64                         `json:"incomeMinor"`
	ExpenseMinor int64                         `json:"expenseMinor"`
	Currency     string                        `json:"currency"`
	ByCategory   []MonthlyCategoryBreakdownDTO `json:"byCategory"`
	GeneratedAt  time.Time                     `json:"generatedAt"`
}
