package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
	"github.com/moon-eye/velune/services/auth-service/internal/repository"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Users          repository.UserRepository
	RefreshTokens  repository.RefreshTokenRepository
	JWTSecret      string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
}

type RegisterInput struct {
	Email        string
	Password     string
	BaseCurrency string
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

type MeResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	BaseCurrency string   `json:"base_currency"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*TokenResponse, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return nil, errs.New("AUTH_VALIDATION_ERROR", "email and password are required", http.StatusBadRequest)
	}

	if s.Users == nil || s.RefreshTokens == nil {
		return nil, errs.New("AUTH_INTERNAL_ERROR", "auth service not wired", http.StatusInternalServerError)
	}

	existing, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errs.New("AUTH_EMAIL_TAKEN", "email already registered", http.StatusConflict)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	baseCur := strings.ToUpper(strings.TrimSpace(in.BaseCurrency))
	if baseCur == "" {
		baseCur = "USD"
	}

	u := &domain.User{
		ID:            uuid.New(),
		Email:         email,
		PasswordHash:  string(hash),
		BaseCurrency: baseCur,
		Version:       1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.Users.Create(ctx, u); err != nil {
		// Handle rare race where email becomes taken between GetByEmail and Create.
		if isUniqueViolation(err) {
			return nil, errs.New("AUTH_EMAIL_TAKEN", "email already registered", http.StatusConflict)
		}
		return nil, err
	}

	return s.issueTokens(ctx, u.ID, u.Email)
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*TokenResponse, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return nil, errs.New("AUTH_VALIDATION_ERROR", "email and password are required", http.StatusBadRequest)
	}
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errs.New("AUTH_INVALID_CREDENTIALS", "invalid credentials", http.StatusUnauthorized)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return nil, errs.New("AUTH_INVALID_CREDENTIALS", "invalid credentials", http.StatusUnauthorized)
	}
	return s.issueTokens(ctx, u.ID, u.Email)
}

func (s *AuthService) Refresh(ctx context.Context, in RefreshInput) (*TokenResponse, error) {
	rt := strings.TrimSpace(in.RefreshToken)
	if rt == "" {
		return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", http.StatusUnauthorized)
	}
	if s.RefreshTokens == nil {
		return nil, errs.New("AUTH_INTERNAL_ERROR", "auth service not wired", http.StatusInternalServerError)
	}

	tokenHash := hashToken(rt)
	existing, err := s.RefreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", http.StatusUnauthorized)
	}

	// Rotate refresh token in-place (old token becomes invalid immediately).
	newRefresh, err := generateOpaqueToken(32)
	if err != nil {
		return nil, err
	}
	newHash := hashToken(newRefresh)
	newExpires := time.Now().UTC().Add(s.refreshTTL())
	if err := s.RefreshTokens.Rotate(ctx, existing.ID, newHash, newExpires); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", http.StatusUnauthorized)
		}
		return nil, err
	}

	user, err := s.Users.GetByID(ctx, existing.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errs.New("AUTH_INVALID_REFRESH_TOKEN", "invalid refresh token", http.StatusUnauthorized)
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
		UserID:        user.ID,
		Email:         user.Email,
		BaseCurrency:  user.BaseCurrency,
	}, nil
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*domain.User, error) {
	claims, err := jwt.Parse(accessToken, s.JWTSecret)
	if err != nil {
		return nil, errs.New("AUTH_UNAUTHORIZED", "unauthorized", http.StatusUnauthorized)
	}
	user, err := s.Users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errs.New("AUTH_UNAUTHORIZED", "unauthorized", http.StatusUnauthorized)
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
	if v := strings.TrimSpace(os.Getenv("REFRESH_TOKEN_TTL")); v != "" {
		if dur, err := time.ParseDuration(v); err == nil && dur > 0 {
			return dur
		}
	}
	return 30 * 24 * time.Hour
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
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
