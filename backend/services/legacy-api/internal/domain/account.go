package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountWallet   AccountType = "wallet"
	AccountBank     AccountType = "bank"
	AccountEMoney   AccountType = "e_money"
	AccountCash     AccountType = "cash"
	AccountCard     AccountType = "card"
)

type Account struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         string
	Type         AccountType
	Currency     string
	BalanceMinor int64 // derived from ledger; cached for reads
	Version      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
