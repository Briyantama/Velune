package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
	"github.com/moon-eye/velune/services/auth-service/internal/repository"
	"github.com/moon-eye/velune/services/auth-service/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	createFn     func(ctx context.Context, u *domain.User) error
	activateFn   func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, u)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.getByIDFn == nil {
		return nil, nil
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn == nil {
		return nil, nil
	}
	return m.getByEmailFn(ctx, email)
}

func (m *mockUserRepo) ActivateAfterOTP(ctx context.Context, userID uuid.UUID) error {
	if m.activateFn == nil {
		return nil
	}
	return m.activateFn(ctx, userID)
}

type rotateCall struct {
	tokenID      uuid.UUID
	newTokenHash string
	newExpiresAt time.Time
}

type mockRefreshRepo struct {
	storeFn      func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	getByHashFn  func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	rotateFn     func(ctx context.Context, tokenID uuid.UUID, newTokenHash string, newExpiresAt time.Time) error
	softDeleteFn func(ctx context.Context, tokenID uuid.UUID) error

	rotateCalls []rotateCall
}

func (m *mockRefreshRepo) Store(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	if m.storeFn == nil {
		return nil
	}
	return m.storeFn(ctx, userID, tokenHash, expiresAt)
}

func (m *mockRefreshRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	if m.getByHashFn == nil {
		return nil, nil
	}
	return m.getByHashFn(ctx, tokenHash)
}

func (m *mockRefreshRepo) Rotate(ctx context.Context, tokenID uuid.UUID, newTokenHash string, newExpiresAt time.Time) error {
	m.rotateCalls = append(m.rotateCalls, rotateCall{
		tokenID:      tokenID,
		newTokenHash: newTokenHash,
		newExpiresAt: newExpiresAt,
	})
	if m.rotateFn == nil {
		return nil
	}
	return m.rotateFn(ctx, tokenID, newTokenHash, newExpiresAt)
}

func (m *mockRefreshRepo) SoftDelete(ctx context.Context, tokenID uuid.UUID) error {
	if m.softDeleteFn == nil {
		return nil
	}
	return m.softDeleteFn(ctx, tokenID)
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

type mockOTPRepo struct {
	getLatestIssuedFn   func(ctx context.Context, userID uuid.UUID) (*time.Time, int, error)
	getLatestUnconsFn   func(ctx context.Context, userID uuid.UUID) (*domain.OTPVerification, error)
	invalidateFn        func(ctx context.Context, userID uuid.UUID) error
	consumeByIDFn       func(ctx context.Context, otpID uuid.UUID) error
	incrementAttemptFn  func(ctx context.Context, otpID uuid.UUID) error
	consumeIfExceededFn func(ctx context.Context, otpID uuid.UUID, maxAttempts int) (int64, error)
	createFn            func(ctx context.Context, otp *domain.OTPVerification) error
}

func (m *mockOTPRepo) Create(ctx context.Context, otp *domain.OTPVerification) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, otp)
}

func (m *mockOTPRepo) GetLatestUnconsumed(ctx context.Context, userID uuid.UUID) (*domain.OTPVerification, error) {
	if m.getLatestUnconsFn == nil {
		return nil, nil
	}
	return m.getLatestUnconsFn(ctx, userID)
}

func (m *mockOTPRepo) InvalidateUnconsumedForUser(ctx context.Context, userID uuid.UUID) error {
	if m.invalidateFn == nil {
		return nil
	}
	return m.invalidateFn(ctx, userID)
}

func (m *mockOTPRepo) ConsumeByID(ctx context.Context, otpID uuid.UUID) error {
	if m.consumeByIDFn == nil {
		return nil
	}
	return m.consumeByIDFn(ctx, otpID)
}

func (m *mockOTPRepo) IncrementAttempt(ctx context.Context, otpID uuid.UUID) error {
	if m.incrementAttemptFn == nil {
		return nil
	}
	return m.incrementAttemptFn(ctx, otpID)
}

func (m *mockOTPRepo) ConsumeIfAttemptsExceeded(ctx context.Context, otpID uuid.UUID, maxAttempts int) (int64, error) {
	if m.consumeIfExceededFn == nil {
		return 0, nil
	}
	return m.consumeIfExceededFn(ctx, otpID, maxAttempts)
}

func (m *mockOTPRepo) GetLatestIssuedMeta(ctx context.Context, userID uuid.UUID) (*time.Time, int, error) {
	if m.getLatestIssuedFn == nil {
		return nil, 0, nil
	}
	return m.getLatestIssuedFn(ctx, userID)
}

type mockOTPSender struct {
	sendFn func(ctx context.Context, toEmail string, otpCode string, expiresAt time.Time) error
}

func (m *mockOTPSender) SendOTP(ctx context.Context, toEmail string, otpCode string, expiresAt time.Time) error {
	if m == nil || m.sendFn == nil {
		return nil
	}
	return m.sendFn(ctx, toEmail, otpCode, expiresAt)
}

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()

	var createdUserID uuid.UUID
	var storedOTPHash string
	otpCodeCh := make(chan string, 1)
	otpRowCreated := make(chan struct{}, 1)

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, u *domain.User) error {
			createdUserID = u.ID
			if u.Email != "user@example.com" {
				t.Fatalf("expected email normalized to lower-case, got %q", u.Email)
			}
			if u.PasswordHash == "" || u.PasswordHash == "password123" {
				t.Fatalf("expected hashed password, got %q", u.PasswordHash)
			}
			if u.Status != "pending" {
				t.Fatalf("expected status pending, got %q", u.Status)
			}
			if u.BaseCurrency != "USD" {
				t.Fatalf("expected base currency USD, got %q", u.BaseCurrency)
			}
			return nil
		},
	}

	otpRepo := &mockOTPRepo{
		getLatestIssuedFn: func(ctx context.Context, userID uuid.UUID) (*time.Time, int, error) {
			return nil, 0, nil
		},
		invalidateFn: func(ctx context.Context, userID uuid.UUID) error {
			return nil
		},
		createFn: func(ctx context.Context, otp *domain.OTPVerification) error {
			storedOTPHash = otp.OtpHash
			if otp.UserID != createdUserID {
				t.Fatalf("expected otp for user %s, got %s", createdUserID, otp.UserID)
			}
			if otp.ResendCount != 0 {
				t.Fatalf("expected initial resendCount=0, got %d", otp.ResendCount)
			}
			otpRowCreated <- struct{}{}
			return nil
		},
	}

	otpSender := &mockOTPSender{
		sendFn: func(ctx context.Context, toEmail string, otpCode string, expiresAt time.Time) error {
			if toEmail != "user@example.com" {
				t.Fatalf("expected otp sent to normalized email, got %q", toEmail)
			}
			otpCodeCh <- otpCode
			return nil
		},
	}

	svc := &usecase.AuthService{
		Log:                  zap.NewNop(),
		Users:                userRepo,
		OTPVerifications:     otpRepo,
		OTPSender:            otpSender,
		OTPValidity:          5 * time.Minute,
		OTPResendCooldown:    30 * time.Second,
		OTPMaxResends:        3,
		OTPMaxVerifyAttempts: 3,
	}

	resp, err := svc.Register(ctx, usecase.RegisterInput{
		Email:        "User@Example.com",
		Password:     "password123",
		BaseCurrency: "usd",
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.Message != "OTP sent" {
		t.Fatalf("expected message 'OTP sent', got %q", resp.Message)
	}

	select {
	case <-otpRowCreated:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for otp verification row creation")
	}

	var otpCode string
	select {
	case otpCode = <-otpCodeCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for otp delivery")
	}

	if storedOTPHash == "" {
		t.Fatalf("expected stored otp hash to be set")
	}
	if storedOTPHash != sha256Hex(otpCode) {
		t.Fatalf("expected stored otp hash to match delivered otp code")
	}
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	accessTTL := 30 * time.Minute
	refreshTTL := 30 * 24 * time.Hour
	userID := uuid.New()

	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}

	var storedRefreshTokenHash string

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			if email != "user@example.com" {
				t.Fatalf("unexpected email: %q", email)
			}
			return &domain.User{
				ID:           userID,
				Email:        email,
				PasswordHash: string(hash),
				BaseCurrency: "USD",
			}, nil
		},
	}

	refreshRepo := &mockRefreshRepo{
		storeFn: func(ctx context.Context, uid uuid.UUID, tokenHash string, expiresAt time.Time) error {
			if uid != userID {
				t.Fatalf("expected uid %s, got %s", userID, uid)
			}
			storedRefreshTokenHash = tokenHash
			return nil
		},
	}

	svc := &usecase.AuthService{
		Users:         userRepo,
		RefreshTokens: refreshRepo,
		JWTSecret:     jwtSecret,
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}

	resp, err := svc.Login(ctx, usecase.LoginInput{
		Email:    "User@Example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("expected access and refresh tokens")
	}

	claims, err := jwt.Parse(resp.AccessToken, jwtSecret)
	if err != nil {
		t.Fatalf("expected valid access token, got error: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected uid %s, got %s", userID, claims.UserID)
	}

	if storedRefreshTokenHash != sha256Hex(resp.RefreshToken) {
		t.Fatalf("expected stored hash to match refresh token hash")
	}
}

func TestLogin_InvalidPassword_ReturnsStableCode(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	accessTTL := 30 * time.Minute
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

	_, err = svc.Login(ctx, usecase.LoginInput{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	var ae *errs.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != "AUTH_INVALID_CREDENTIALS" {
		t.Fatalf("expected code AUTH_INVALID_CREDENTIALS, got %q", ae.Code)
	}
}

func TestRefresh_RotatesAndReturnsNewTokens(t *testing.T) {
	ctx := context.Background()
	jwtSecret := "test-secret"
	accessTTL := 1 * time.Hour
	refreshTTL := 30 * 24 * time.Hour

	userID := uuid.New()
	existingTokenID := uuid.New()

	oldRefreshToken := "old-refresh-token"
	oldHash := sha256Hex(oldRefreshToken)

	newUser := &domain.User{
		ID:           userID,
		Email:        "user@example.com",
		PasswordHash: "irrelevant",
		BaseCurrency: "USD",
	}

	var rotatedTokenHash string

	refreshRepo := &mockRefreshRepo{
		getByHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			if tokenHash != oldHash {
				t.Fatalf("expected token hash %s, got %s", oldHash, tokenHash)
			}
			return &domain.RefreshToken{
				ID:        existingTokenID,
				UserID:    userID,
				TokenHash: oldHash,
				ExpiresAt: time.Now().UTC().Add(refreshTTL / 2),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				Version:   1,
			}, nil
		},
		rotateFn: func(ctx context.Context, tokenID uuid.UUID, newTokenHash string, newExpiresAt time.Time) error {
			if tokenID != existingTokenID {
				t.Fatalf("expected rotate for token %s, got %s", existingTokenID, tokenID)
			}
			rotatedTokenHash = newTokenHash
			d := time.Until(newExpiresAt)
			if d < refreshTTL-time.Minute || d > refreshTTL+time.Minute {
				t.Fatalf("unexpected new refresh ttl: %s (expected ~%s)", d, refreshTTL)
			}
			return nil
		},
	}

	userRepo := &mockUserRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id != userID {
				t.Fatalf("expected uid %s, got %s", userID, id)
			}
			return newUser, nil
		},
	}

	svc := &usecase.AuthService{
		Users:         userRepo,
		RefreshTokens: refreshRepo,
		JWTSecret:     jwtSecret,
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}

	resp, err := svc.Refresh(ctx, usecase.RefreshInput{RefreshToken: oldRefreshToken})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("expected new tokens")
	}
	if len(refreshRepo.rotateCalls) != 1 {
		t.Fatalf("expected exactly 1 rotate call, got %d", len(refreshRepo.rotateCalls))
	}

	claims, err := jwt.Parse(resp.AccessToken, jwtSecret)
	if err != nil {
		t.Fatalf("expected valid access token, got error: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected uid %s, got %s", userID, claims.UserID)
	}

	if rotatedTokenHash != sha256Hex(resp.RefreshToken) {
		t.Fatalf("expected rotated token hash to match returned refresh token hash")
	}
}

func TestRefresh_InvalidToken_ReturnsStableCode(t *testing.T) {
	ctx := context.Background()
	svc := &usecase.AuthService{
		JWTSecret:  "test-secret",
		AccessTTL:  1 * time.Hour,
		RefreshTTL: 30 * 24 * time.Hour,
		Users:      &mockUserRepo{},
		RefreshTokens: &mockRefreshRepo{
			getByHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
				return nil, nil
			},
		},
	}

	_, err := svc.Refresh(ctx, usecase.RefreshInput{RefreshToken: "unknown"})
	if err == nil {
		t.Fatalf("expected error")
	}
	var ae *errs.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != "AUTH_INVALID_REFRESH_TOKEN" {
		t.Fatalf("expected code AUTH_INVALID_REFRESH_TOKEN, got %q", ae.Code)
	}
}

// Compile-time interface checks.
var _ repository.UserRepository = (*mockUserRepo)(nil)
var _ repository.RefreshTokenRepository = (*mockRefreshRepo)(nil)
var _ = pgconn.PgError{} // avoid unused imports if build tags change
