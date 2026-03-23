package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	BaseCurrency string // default ISO 4217 for new accounts / UX
	Version      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
