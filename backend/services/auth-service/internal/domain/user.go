package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the auth bounded context identity aggregate (owned only by auth-service).
type User struct {
	ID              uuid.UUID
	Email           string
	PasswordHash    string
	BaseCurrency    string
	Status          string
	EmailVerifiedAt *time.Time
	DisplayName     *string
	Version         int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
