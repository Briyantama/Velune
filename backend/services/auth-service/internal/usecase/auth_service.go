package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
	"github.com/moon-eye/velune/services/auth-service/internal/repository"
	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/helper"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/jwt"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/stringx"
	"go.uber.org/zap"
)

type AuthService struct {
	Log               *zap.Logger
	Users             repository.UserRepository
	RefreshTokens     repository.RefreshTokenRepository
	OTPVerifications  repository.OTPVerificationRepository
	ProvisioningState repository.ProvisioningStateRepository
	OTPSender         OTPSender

	EventPublisher EventPublisher

	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration

	OTPValidity          time.Duration
	OTPResendCooldown    time.Duration
	OTPMaxResends        int
	OTPMaxVerifyAttempts int
}

type EventPublisher interface {
	Publish(ctx context.Context, env contracts.EventEnvelope) error
}

type RegisterInput struct {
	Email        string
	Password     string
	BaseCurrency string
	DisplayName  *string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type OTPRegisterResponse struct {
	Message string `json:"message"`
}

type VerifyOTPInput struct {
	Email   string
	OTPCode string
}

type ResendOTPInput struct {
	Email string
}

type MeResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	BaseCurrency string    `json:"base_currency"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*OTPRegisterResponse, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return nil, errs.New("AUTH_VALIDATION_ERROR", "email and password are required", constx.StatusBadRequest)
	}

	if s.Users == nil || s.OTPVerifications == nil || s.OTPSender == nil {
		return nil, errs.New("AUTH_INTERNAL_ERROR", "auth service not wired", constx.StatusInternalServerError)
	}

	existing, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	// Do not reveal whether email exists: we always return `{message:\"OTP sent\"}` for valid inputs.

	hash, err := jwt.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	baseCur := stringx.Upper(stringx.TrimSpace(in.BaseCurrency))

	u := existing
	if u == nil {
		u = &domain.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: hash,
			BaseCurrency: baseCur,
			Status:       "pending",
			DisplayName:  in.DisplayName,
			Version:      1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.Users.Create(ctx, u); err != nil {
			// Handle rare race where email becomes taken between GetByEmail and Create.
			if isUniqueViolation(err) {
				u, err = s.Users.GetByEmail(ctx, email)
				if err != nil {
					return nil, err
				}
				if u == nil {
					return nil, errs.New("AUTH_INTERNAL_ERROR", "auth user creation failed", constx.StatusInternalServerError)
				}
			} else {
				return nil, err
			}
		}
	}

	code, codeHash, err := generateOTP()
	if err != nil {
		return nil, err
	}

	// Determine resend_count (register calls count as an issuance).
	lastIssuedAt, lastResendCount, err := s.OTPVerifications.GetLatestIssuedMeta(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	resendCount := 0
	if lastIssuedAt != nil {
		resendCount = lastResendCount + 1
	}

	expiresAt := now.Add(s.OTPValidity)

	// Invalidate any existing OTP for the user before issuing a new one.
	_ = s.OTPVerifications.InvalidateUnconsumedForUser(ctx, u.ID)
	if err := s.OTPVerifications.Create(ctx, &domain.OTPVerification{
		ID:                 uuid.New(),
		UserID:             u.ID,
		Email:              u.Email,
		OtpHash:            codeHash,
		ExpiresAt:          expiresAt,
		ConsumedAt:         nil,
		ResendCount:        resendCount,
		VerifyAttemptCount: 0,
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}); err != nil {
		return nil, err
	}

	// Send asynchronously; if delivery fails we do not block registration completion.
	if s.Log != nil {
		go func() {
			sendCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := s.OTPSender.SendOTP(sendCtx, u.Email, code, expiresAt); err != nil {
				s.Log.Warn("otp_send_failed", append(sharedlog.FieldsFromContext(ctx), zap.Error(err))...)
			}
		}()
	} else {
		go func() {
			sendCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_ = s.OTPSender.SendOTP(sendCtx, u.Email, code, expiresAt)
		}()
	}

	return &OTPRegisterResponse{Message: "OTP sent"}, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*TokenResponse, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return nil, errs.New("AUTH_VALIDATION_ERROR", "email and password are required", constx.StatusBadRequest)
	}
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errs.New("AUTH_INVALID_CREDENTIALS", "invalid credentials", constx.StatusUnauthorized)
	}
	if !jwt.VerifyPassword(in.Password, u.PasswordHash) {
		return nil, errs.New("AUTH_INVALID_CREDENTIALS", "invalid credentials", constx.StatusUnauthorized)
	}
	// Backward compatibility: older users might not have status filled. Treat empty as active.
	status := u.Status
	if status == "" {
		status = "active"
	}
	if status != "active" {
		return nil, errs.ErrUnauthorized
	}
	resp, err := s.issueTokens(ctx, u.ID, u.Email)
	if err != nil {
		return nil, err
	}

	// Kick off default account provisioning for first-time users.
	// This must not block login; broker publish failures should be retried on subsequent logins.
	if s.ProvisioningState != nil && s.EventPublisher != nil {
		go func() {
			at, err := s.ProvisioningState.GetAccountProvisionedAt(ctx, u.ID)
			if err != nil {
				if s.Log != nil {
					s.Log.Warn("provisioning_state_check_failed", append(sharedlog.FieldsFromContext(ctx), zap.Error(err))...)
				}
				return
			}
			if at != nil {
				return
			}

			payload := contracts.UserFirstLoginProvisionRequested{
				UserID:       u.ID,
				Email:        u.Email,
				DisplayName:  u.DisplayName,
				BaseCurrency: u.BaseCurrency,
				OccurredAt:   time.Now().UTC(),
			}
			payloadBytes, err := helper.ToJSONMarshal(payload)
			if err != nil {
				return
			}

			env := contracts.EventEnvelope{
				EventID:     uuid.New(),
				EventType:   contracts.EventUserFirstLoginProvisionRequested,
				Version:     "v1",
				Source:      "auth-service",
				OccurredAt:  time.Now().UTC(),
				UserID:      &u.ID,
				Idempotency: fmt.Sprintf("provision_requested:%s", u.ID.String()),
				Payload:     payloadBytes,
			}
			if cid, ok := httpx.CorrelationID(ctx); ok {
				env.CorrelationID = cid
			}

			_ = s.EventPublisher.Publish(context.Background(), env)
		}()
	}

	return resp, nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, in VerifyOTPInput) error {
	email := normalizeEmail(in.Email)
	code := stringx.TrimSpace(in.OTPCode)
	if email == "" || code == "" {
		return errs.New("AUTH_VALIDATION_ERROR", "email and otp are required", constx.StatusBadRequest)
	}
	if s.Users == nil || s.OTPVerifications == nil {
		return errs.New("AUTH_INTERNAL_ERROR", "auth service not wired", constx.StatusInternalServerError)
	}

	user, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return errs.New("AUTH_INVALID_OTP", "invalid otp", constx.StatusUnauthorized)
	}

	latest, err := s.OTPVerifications.GetLatestUnconsumed(ctx, user.ID)
	if err != nil {
		return err
	}
	if latest == nil {
		return errs.New("AUTH_INVALID_OTP", "invalid otp", constx.StatusUnauthorized)
	}

	now := time.Now().UTC()
	if now.After(latest.ExpiresAt) {
		_ = s.OTPVerifications.ConsumeByID(ctx, latest.ID)
		return errs.New("AUTH_INVALID_OTP", "invalid otp", constx.StatusUnauthorized)
	}

	_, expectedHash, err := generateOTPFromCode(code)
	if err != nil {
		return err
	}
	if !constantTimeStringEq(expectedHash, latest.OtpHash) {
		// Mismatch: increment attempt counter and optionally invalidate OTP when attempts are exceeded.
		_ = s.OTPVerifications.IncrementAttempt(ctx, latest.ID)
		_, _ = s.OTPVerifications.ConsumeIfAttemptsExceeded(ctx, latest.ID, s.OTPMaxVerifyAttempts)
		return errs.New("AUTH_INVALID_OTP", "invalid otp", constx.StatusUnauthorized)
	}

	// OTP is valid: consume and activate user.
	if err := s.OTPVerifications.ConsumeByID(ctx, latest.ID); err != nil {
		return err
	}
	if err := s.Users.ActivateAfterOTP(ctx, user.ID); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ResendOTP(ctx context.Context, in ResendOTPInput) error {
	email := normalizeEmail(in.Email)
	if email == "" {
		return errs.New("AUTH_VALIDATION_ERROR", "email is required", constx.StatusBadRequest)
	}
	if s.Users == nil || s.OTPVerifications == nil || s.OTPSender == nil {
		return errs.New("AUTH_INTERNAL_ERROR", "auth service not wired", constx.StatusInternalServerError)
	}

	user, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	// Enumeration safety: for missing users, act like success.
	if user == nil {
		return nil
	}

	lastIssuedAt, lastResendCount, err := s.OTPVerifications.GetLatestIssuedMeta(ctx, user.ID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	// Cooldown: require some time between OTP issuances.
	if lastIssuedAt != nil && s.OTPResendCooldown > 0 && now.Sub(*lastIssuedAt) < s.OTPResendCooldown {
		return errs.New("AUTH_OTP_RESEND_RATE_LIMIT", "otp resend not allowed yet", constx.StatusTooManyRequests)
	}
	// Max resend attempts (does not count the initial issuance when lastIssuedAt is nil).
	nextResendCount := 0
	if lastIssuedAt != nil {
		nextResendCount = lastResendCount + 1
		if s.OTPMaxResends >= 0 && lastResendCount >= s.OTPMaxResends {
			return errs.New("AUTH_OTP_RESEND_LIMIT_REACHED", "otp resend limit reached", constx.StatusTooManyRequests)
		}
	}

	code, codeHash, err := generateOTP()
	if err != nil {
		return err
	}
	expiresAt := now.Add(s.OTPValidity)

	_ = s.OTPVerifications.InvalidateUnconsumedForUser(ctx, user.ID)
	if err := s.OTPVerifications.Create(ctx, &domain.OTPVerification{
		ID:                 uuid.New(),
		UserID:             user.ID,
		Email:              user.Email,
		OtpHash:            codeHash,
		ExpiresAt:          expiresAt,
		ConsumedAt:         nil,
		ResendCount:        nextResendCount,
		VerifyAttemptCount: 0,
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}); err != nil {
		return err
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.OTPSender.SendOTP(sendCtx, user.Email, code, expiresAt); err != nil && s.Log != nil {
			s.Log.Warn("otp_send_failed", append(sharedlog.FieldsFromContext(ctx), zap.Error(err))...)
		}
	}()

	return nil
}

func (s *AuthService) Refresh(ctx context.Context, in RefreshInput) (*TokenResponse, error) {
	rt := stringx.TrimSpace(in.RefreshToken)
	if rt == "" {
		return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", constx.StatusUnauthorized)
	}
	if s.RefreshTokens == nil {
		return nil, errs.New("AUTH_INTERNAL_ERROR", "auth service not wired", constx.StatusInternalServerError)
	}

	tokenHash := hashToken(rt)
	existing, err := s.RefreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", constx.StatusUnauthorized)
	}

	// Rotate refresh token in-place (old token becomes invalid immediately).
	newRefresh, err := generateOpaqueToken(32)
	if err != nil {
		return nil, err
	}
	newHash := hashToken(newRefresh)
	newExpires := time.Now().UTC().Add(s.refreshTTL())
	if err := s.RefreshTokens.Rotate(ctx, existing.ID, newHash, newExpires); err != nil {
		if errors.Is(err, errs.ErrRefreshToken) {
			return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", constx.StatusUnauthorized)
		}
		return nil, err
	}

	user, err := s.Users.GetByID(ctx, existing.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", constx.StatusUnauthorized)
	}
	accessResp, err := s.issueAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	accessResp.RefreshToken = newRefresh
	accessResp.ExpiresIn = int64(s.AccessTTL.Seconds())
	return accessResp, nil
}

func (s *AuthService) Me(ctx context.Context, accessToken string) (*MeResponse, error) {
	user, err := s.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return &MeResponse{
		UserID:       user.ID,
		Email:        user.Email,
		BaseCurrency: user.BaseCurrency,
	}, nil
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*domain.User, error) {
	claims, err := jwt.Parse(accessToken, s.JWTSecret)
	if err != nil {
		return nil, errs.New("AUTH_UNAUTHORIZED", "unauthorized", constx.StatusUnauthorized)
	}
	user, err := s.Users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errs.New("AUTH_UNAUTHORIZED", "unauthorized", constx.StatusUnauthorized)
	}
	return user, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID, email string) (*TokenResponse, error) {
	resp, err := s.issueAccessToken(userID, email)
	if err != nil {
		return nil, err
	}
	rt, err := generateOpaqueToken(32)
	if err != nil {
		return nil, err
	}
	rtHash := hashToken(rt)
	expires := time.Now().UTC().Add(s.refreshTTL())
	if err := s.RefreshTokens.Store(ctx, userID, rtHash, expires); err != nil {
		return nil, err
	}
	resp.RefreshToken = rt
	return resp, nil
}

func (s *AuthService) issueAccessToken(userID uuid.UUID, email string) (*TokenResponse, error) {
	access, err := jwt.Sign(userID, email, s.JWTSecret, s.AccessTTL)
	if err != nil {
		return nil, err
	}
	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: "",
		ExpiresIn:    int64(s.AccessTTL.Seconds()),
	}, nil
}

func (s *AuthService) refreshTTL() time.Duration {
	if s.RefreshTTL > 0 {
		return s.RefreshTTL
	}
	// Safe default: 30 days (overridable via REFRESH_TOKEN_TTL).
	if v := stringx.TrimSpace(os.Getenv("REFRESH_TOKEN_TTL")); v != "" {
		if dur, err := time.ParseDuration(v); err == nil && dur > 0 {
			return dur
		}
	}
	return 30 * 24 * time.Hour
}

func normalizeEmail(email string) string {
	return stringx.Lower(stringx.TrimSpace(email))
}

func generateOpaqueToken(nbytes int) (string, error) {
	b := make([]byte, nbytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe, opaque string (no padding).
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func generateOTP() (code string, codeHash string, err error) {
	// 6-digit, zero-padded numeric OTP.
	limit := big.NewInt(1_000_000) // 0..999999
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", "", err
	}
	code = fmt.Sprintf("%06d", n.Int64())
	return code, otpHash(code), nil
}

func generateOTPFromCode(code string) (string, string, error) {
	if len(code) != 6 {
		return "", "", errs.New("AUTH_VALIDATION_ERROR", "invalid otp", constx.StatusBadRequest)
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return "", "", errs.New("AUTH_VALIDATION_ERROR", "invalid otp", constx.StatusBadRequest)
		}
	}
	return code, otpHash(code), nil
}

func otpHash(otpCode string) string {
	sum := sha256.Sum256([]byte(otpCode))
	return hex.EncodeToString(sum[:])
}

func constantTimeStringEq(a, b string) bool {
	// Both OTP hashes are fixed-length SHA256 hex strings (64 chars).
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
