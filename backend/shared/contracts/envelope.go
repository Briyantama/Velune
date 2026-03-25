// Package contracts holds cross-service DTOs and event shapes. Keep stable for API versioning.
package contracts

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ErrorResponse is the standard JSON error body for REST APIs.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PagedMeta accompanies list responses.
type PagedMeta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// UserRef identifies a user across services (no PII beyond id).
type UserRef struct {
	ID uuid.UUID `json:"id"`
}

// MoneyAmount is a portable money representation (minor units + ISO currency).
type MoneyAmount struct {
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
}

// EventEnvelope is the shared broker message envelope across services.
type EventEnvelope struct {
	EventID       uuid.UUID       `json:"eventId"`
	EventType     string          `json:"eventType"`
	Version       string          `json:"version,omitempty"`
	Source        string          `json:"source"`
	OccurredAt    time.Time       `json:"occurredAt"`
	UserID        *uuid.UUID      `json:"userId,omitempty"`
	Idempotency   string          `json:"idempotencyKey,omitempty"`
	CorrelationID string          `json:"correlationId,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}
