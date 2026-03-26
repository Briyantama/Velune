package contracts

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventUserFirstLoginProvisionRequested = "user.first_login.provision_requested"
	EventUserFirstLoginProvisionCompleted = "user.first_login.provision_completed"
)

// UserFirstLoginProvisionRequested is emitted when an active user logs in for the first time
// and default financial accounts must be provisioned by transaction-service.
type UserFirstLoginProvisionRequested struct {
	UserID       uuid.UUID `json:"userId"`
	Email        string    `json:"email"`
	DisplayName  *string   `json:"displayName,omitempty"`
	BaseCurrency string    `json:"baseCurrency"`
	OccurredAt   time.Time `json:"occurredAt"`
}

// UserFirstLoginProvisionCompleted is emitted after transaction-service provisions the accounts.
type UserFirstLoginProvisionCompleted struct {
	UserID     uuid.UUID `json:"userId"`
	OccurredAt time.Time `json:"occurredAt"`
}
