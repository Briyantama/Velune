package email

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// StubOtpSender is used in dev/test when SMTP is not configured.
// It does not deliver email.
type StubOtpSender struct {
	Log   *zap.Logger
	From  string
	Allow func(toEmail string, expiresAt time.Time) error
}

func (s *StubOtpSender) SendOTP(ctx context.Context, toEmail string, otpCode string, expiresAt time.Time) error {
	if s == nil {
		return fmt.Errorf("otp sender not configured")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if s.Log != nil {
		// Never log otpCode.
		s.Log.Info("otp_send_stub", zap.String("to", toEmail), zap.Time("expires_at", expiresAt.UTC()))
	}
	if s.Allow != nil {
		return s.Allow(toEmail, expiresAt)
	}
	return nil
}
