package domain

import (
	"time"

	"github.com/google/uuid"
)

// OTPVerification stores a hashed one-time password for email verification.
// It is write-once (consumed_at marks it invalid) to support idempotency and replay safety.
type OTPVerification struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Email              string
	OtpHash            string
	ExpiresAt          time.Time
	ConsumedAt         *time.Time
	ResendCount        int
	VerifyAttemptCount int
	Version            int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}
