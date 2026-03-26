package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProvisioningState tracks whether transaction-service has provisioned
// default financial accounts for a given auth user.
type ProvisioningState struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	AccountProvisionedAt *time.Time
	Version              int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}
