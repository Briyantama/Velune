package domain

import (
	"time"

	"github.com/google/uuid"
)

type MonthlyCategoryBreakdown struct {
	CategoryID   *uuid.UUID
	CategoryName string
	TotalMinor   int64
	Currency     string
}

type MonthlyReport struct {
	UserID       uuid.UUID
	Year         int
	Month        int
	IncomeMinor  int64
	ExpenseMinor int64
	Currency     string
	ByCategory   []MonthlyCategoryBreakdown
	GeneratedAt  time.Time
}
