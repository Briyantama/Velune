package usecase

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
	"github.com/moon-eye/velune/services/auth-service/internal/usecase"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
	"golang.org/x/crypto/bcrypt"
)

type mockProvisioningStateRepo struct {
	getFn  func(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	markFn func(ctx context.Context, userID uuid.UUID, at time.Time) error
}

func (m *mockProvisioningStateRepo) GetAccountProvisionedAt(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx, userID)
}

func (m *mockProvisioningStateRepo) MarkAccountProvisionedAt(ctx context.Context, userID uuid.UUID, at time.Time) error {
	if m.markFn == nil {
		return nil
	}
	return m.markFn(ctx, userID, at)
}

type mockEventPublisher struct {
	publishFn func(ctx context.Context, env contracts.EventEnvelope) error
	calls     int32
}

func (m *mockEventPublisher) Publish(ctx context.Context, env contracts.EventEnvelope) error {
	atomic.AddInt32(&m.calls, 1)
	if m.publishFn == nil {
		return nil
	}
	return m.publishFn(ctx, env)
}

func TestVerifyOTP_Success_ActivatesUser(t *testing.T) {
	ctx := context.Background()
	email := "user@example.com"
	userID := uuid.New()
	otpID := uuid.New()
	otpCode := "123456"
	now := time.Now().UTC().Add(-1 * time.Minute)
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, e string) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        e,
				PasswordHash: "ignored",
				BaseCurrency: "USD",
				Status:       "pending",
			}, nil
		},
		activateFn: func(ctx context.Context, uid uuid.UUID) error {
			if uid != userID {
				t.Fatalf("expected activate for %s, got %s", userID, uid)
			}
			return nil
		},
	}

	otpRepo := &mockOTPRepo{
		getLatestUnconsFn: func(ctx context.Context, uid uuid.UUID) (*domain.OTPVerification, error) {
			if uid != userID {
				t.Fatalf("expected otp for %s, got %s", userID, uid)
			}
			return &domain.OTPVerification{
				ID:                 otpID,
				UserID:             userID,
				Email:              email,
				OtpHash:            sha256Hex(otpCode),
				ExpiresAt:          expiresAt,
				ConsumedAt:         nil,
				ResendCount:        0,
				VerifyAttemptCount: 0,
			}, nil
		},
		consumeByIDFn: func(ctx context.Context, id uuid.UUID) error {
			if id != otpID {
				t.Fatalf("expected consume for %s, got %s", otpID, id)
			}
			return nil
		},
	}

	svc := &usecase.AuthService{
		Log:                  nil,
		Users:                userRepo,
		OTPVerifications:     otpRepo,
		OTPMaxVerifyAttempts: 3,
		OTPValidity:          5 * time.Minute,
		OTPResendCooldown:    30 * time.Second,
	}

	if err := svc.VerifyOTP(ctx, usecase.VerifyOTPInput{Email: email, OTPCode: otpCode}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	_ = now
}

func TestVerifyOTP_InvalidOtp_IncrementsAttemptAndReturnsStableError(t *testing.T) {
	ctx := context.Background()
	email := "user@example.com"
	userID := uuid.New()
	otpID := uuid.New()
	otpCode := "123456"

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, e string) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        e,
				BaseCurrency: "USD",
				Status:       "pending",
			}, nil
		},
	}

	var incremented bool
	otpRepo := &mockOTPRepo{
		getLatestUnconsFn: func(ctx context.Context, uid uuid.UUID) (*domain.OTPVerification, error) {
			return &domain.OTPVerification{
				ID:                 otpID,
				UserID:             userID,
				OtpHash:            sha256Hex(otpCode),
				ExpiresAt:          time.Now().UTC().Add(5 * time.Minute),
				ConsumedAt:         nil,
				ResendCount:        0,
				VerifyAttemptCount: 0,
			}, nil
		},
		incrementAttemptFn: func(ctx context.Context, id uuid.UUID) error {
			if id != otpID {
				t.Fatalf("expected attempt increment for %s, got %s", otpID, id)
			}
			incremented = true
			return nil
		},
		consumeIfExceededFn: func(ctx context.Context, id uuid.UUID, maxAttempts int) (int64, error) {
			if id != otpID {
				t.Fatalf("expected consumeIfAttemptsExceeded for %s, got %s", otpID, id)
			}
			if maxAttempts != 3 {
				t.Fatalf("expected maxAttempts=3, got %d", maxAttempts)
			}
			return 1, nil
		},
	}

	svc := &usecase.AuthService{
		Users:                userRepo,
		OTPVerifications:     otpRepo,
		OTPMaxVerifyAttempts: 3,
		OTPValidity:          5 * time.Minute,
		OTPResendCooldown:    30 * time.Second,
	}

	err := svc.VerifyOTP(ctx, usecase.VerifyOTPInput{Email: email, OTPCode: "000000"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !incremented {
		t.Fatalf("expected attempt increment")
	}

	var ae *errs.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != "AUTH_INVALID_OTP" {
		t.Fatalf("expected AUTH_INVALID_OTP, got %q", ae.Code)
	}
}

func TestResendOTP_RateLimited(t *testing.T) {
	ctx := context.Background()
	email := "user@example.com"
	userID := uuid.New()
	now := time.Now().UTC()

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, e string) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        e,
				BaseCurrency: "USD",
				Status:       "pending",
			}, nil
		},
	}

	otpRepo := &mockOTPRepo{
		getLatestIssuedFn: func(ctx context.Context, uid uuid.UUID) (*time.Time, int, error) {
			lastIssuedAt := now.Add(-10 * time.Second)
			return &lastIssuedAt, 0, nil
		},
	}

	svc := &usecase.AuthService{
		Users:                userRepo,
		OTPVerifications:     otpRepo,
		OTPSender:            &mockOTPSender{sendFn: func(ctx context.Context, toEmail, otpCode string, expiresAt time.Time) error { return nil }},
		OTPResendCooldown:    30 * time.Second,
		OTPMaxResends:        3,
		OTPValidity:          5 * time.Minute,
		OTPMaxVerifyAttempts: 3,
	}

	err := svc.ResendOTP(ctx, usecase.ResendOTPInput{Email: email})
	if err == nil {
		t.Fatalf("expected error")
	}
	var ae *errs.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != "AUTH_OTP_RESEND_RATE_LIMIT" {
		t.Fatalf("expected AUTH_OTP_RESEND_RATE_LIMIT, got %q", ae.Code)
	}
}

func TestResendOTP_MaxResendsExceeded(t *testing.T) {
	ctx := context.Background()
	email := "user@example.com"
	userID := uuid.New()
	now := time.Now().UTC()

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, e string) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        e,
				BaseCurrency: "USD",
				Status:       "pending",
			}, nil
		},
	}

	otpRepo := &mockOTPRepo{
		getLatestIssuedFn: func(ctx context.Context, uid uuid.UUID) (*time.Time, int, error) {
			lastIssuedAt := now.Add(-2 * time.Minute)
			// lastResendCount == max => the *next* resend must be blocked.
			return &lastIssuedAt, 3, nil
		},
	}

	svc := &usecase.AuthService{
		Users:                userRepo,
		OTPVerifications:     otpRepo,
		OTPSender:            &mockOTPSender{sendFn: func(ctx context.Context, toEmail, otpCode string, expiresAt time.Time) error { return nil }},
		OTPResendCooldown:    30 * time.Second,
		OTPMaxResends:        3,
		OTPValidity:          5 * time.Minute,
		OTPMaxVerifyAttempts: 3,
	}

	err := svc.ResendOTP(ctx, usecase.ResendOTPInput{Email: email})
	if err == nil {
		t.Fatalf("expected error")
	}
	var ae *errs.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != "AUTH_OTP_RESEND_LIMIT_REACHED" {
		t.Fatalf("expected AUTH_OTP_RESEND_LIMIT_REACHED, got %q", ae.Code)
	}
}

func TestLogin_Gating_PendingUsersBlocked(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	accessTTL := 5 * time.Minute
	refreshTTL := 30 * 24 * time.Hour
	userID := uuid.New()

	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        email,
				PasswordHash: string(hash),
				BaseCurrency: "USD",
				Status:       "pending",
			}, nil
		},
	}
	refreshRepo := &mockRefreshRepo{}

	svc := &usecase.AuthService{
		Users:         userRepo,
		RefreshTokens: refreshRepo,
		JWTSecret:     jwtSecret,
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}

	_, err = svc.Login(ctx, usecase.LoginInput{Email: "user@example.com", Password: password})
	if err == nil {
		t.Fatalf("expected error")
	}
	var ae *errs.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != errs.ErrUnauthorized.Code {
		t.Fatalf("expected UNAUTHORIZED, got %q", ae.Code)
	}
}

func TestLogin_ProvisioningRequest_PublishedOnce(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	accessTTL := 5 * time.Minute
	refreshTTL := 30 * 24 * time.Hour
	userID := uuid.New()

	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        email,
				PasswordHash: string(hash),
				BaseCurrency: "USD",
				Status:       "active",
			}, nil
		},
	}

	refreshRepo := &mockRefreshRepo{
		storeFn: func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
			return nil
		},
	}

	var publishCalls int32
	publishCh := make(chan struct{}, 1)
	pub := &mockEventPublisher{
		publishFn: func(ctx context.Context, env contracts.EventEnvelope) error {
			atomic.AddInt32(&publishCalls, 1)
			publishCh <- struct{}{}
			return nil
		},
	}

	callCount := int32(0)
	provRepo := &mockProvisioningStateRepo{
		getFn: func(ctx context.Context, uid uuid.UUID) (*time.Time, error) {
			n := atomic.AddInt32(&callCount, 1)
			if n == 1 {
				return nil, nil
			}
			t := time.Now().UTC()
			return &t, nil
		},
	}

	svc := &usecase.AuthService{
		Users:             userRepo,
		RefreshTokens:     refreshRepo,
		JWTSecret:         jwtSecret,
		AccessTTL:         accessTTL,
		RefreshTTL:        refreshTTL,
		ProvisioningState: provRepo,
		EventPublisher:    pub,
	}

	// First login should trigger provisioning request.
	if _, err := svc.Login(ctx, usecase.LoginInput{Email: "user@example.com", Password: password}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	select {
	case <-publishCh:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for provisioning publish")
	}

	// Second login should not publish again (marker now set).
	if _, err := svc.Login(ctx, usecase.LoginInput{Email: "user@example.com", Password: password}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&publishCalls) != 1 {
		t.Fatalf("expected publish once, got %d", publishCalls)
	}
}
