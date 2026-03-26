package usecase

import (
	"context"
	"time"
)

// OTPSender delivers a one-time password to a destination email.
// Implementations must never log secrets (OTP code).
type OTPSender interface {
	SendOTP(ctx context.Context, toEmail string, otpCode string, expiresAt time.Time) error
}
