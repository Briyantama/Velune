package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
)

// OTPVerificationRepository persists OTP verification attempts.
// OTP hashes are stored at rest, never plaintext.
type OTPVerificationRepository interface {
	Create(ctx context.Context, otp *domain.OTPVerification) error
	GetLatestUnconsumed(ctx context.Context, userID uuid.UUID) (*domain.OTPVerification, error)
	InvalidateUnconsumedForUser(ctx context.Context, userID uuid.UUID) error
	ConsumeByID(ctx context.Context, otpID uuid.UUID) error
	IncrementAttempt(ctx context.Context, otpID uuid.UUID) error
	ConsumeIfAttemptsExceeded(ctx context.Context, otpID uuid.UUID, maxAttempts int) (int64, error)
	GetLatestIssuedMeta(ctx context.Context, userID uuid.UUID) (lastIssuedAt *time.Time, lastResendCount int, err error)
}
