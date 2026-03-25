package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/config"
	"github.com/moon-eye/velune/services/legacy-api/internal/domain"
	"github.com/moon-eye/velune/services/legacy-api/internal/repository"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Users  repository.UserRepository
	Cfg    *config.Config
	Logger *zap.Logger
}

type RegisterInput struct {
	Email        string `validate:"required,email"`
	Password     string `validate:"required,min=8,max=72"`
	BaseCurrency string `validate:"required,len=3"`
}

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type AuthToken struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
	TokenType   string `json:"tokenType"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthToken, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	exists, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, errs.New("EMAIL_TAKEN", "email already registered",constx.StatusConflict)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	u := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		BaseCurrency: strings.ToUpper(in.BaseCurrency),
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.Users.Create(ctx, u); err != nil {
		return nil, err
	}
	return s.issueToken(u)
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthToken, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	u, err := s.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errs.New("INVALID_CREDENTIALS", "invalid email or password",constx.StatusUnauthorized)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errs.New("INVALID_CREDENTIALS", "invalid email or password",constx.StatusUnauthorized)
		}
		return nil, err
	}
	return s.issueToken(u)
}

func (s *AuthService) issueToken(u *domain.User) (*AuthToken, error) {
	if s.Cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is not configured")
	}
	tok, err := jwt.Sign(u.ID, u.Email, s.Cfg.JWTSecret, s.Cfg.JWTExpiry)
	if err != nil {
		return nil, err
	}
	return &AuthToken{
		AccessToken: tok,
		ExpiresIn:   int64(s.Cfg.JWTExpiry.Seconds()),
		TokenType:   "Bearer",
	}, nil
}
