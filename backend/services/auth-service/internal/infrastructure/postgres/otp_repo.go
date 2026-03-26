package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
	"github.com/moon-eye/velune/services/auth-service/internal/repository"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
	stringx "github.com/moon-eye/velune/shared/stringx"
)

type OTPVerificationRepo struct {
	s *Store
}

func NewOTPVerificationRepo(s *Store) repository.OTPVerificationRepository {
	return &OTPVerificationRepo{s: s}
}

func (r *OTPVerificationRepo) Create(ctx context.Context, otp *domain.OTPVerification) error {
	return r.s.Queries.OTPInsert(ctx, db.OTPInsertParams{
		ID:                 helper.ToPgUUID(otp.ID),
		UserID:             helper.ToPgUUID(otp.UserID),
		Email:              stringx.Lower(otp.Email),
		OtpHash:            otp.OtpHash,
		ExpiresAt:          helper.ToPgTS(otp.ExpiresAt),
		ConsumedAt:         helper.ToPgTSPtr(otp.ConsumedAt),
		ResendCount:        int32(otp.ResendCount),
		VerifyAttemptCount: int32(otp.VerifyAttemptCount),
		Version:            otp.Version,
		CreatedAt:          helper.ToPgTS(otp.CreatedAt),
		UpdatedAt:          helper.ToPgTS(otp.UpdatedAt),
	})
}

func (r *OTPVerificationRepo) GetLatestUnconsumed(ctx context.Context, userID uuid.UUID) (*domain.OTPVerification, error) {
	row, err := r.s.Queries.OTPGetLatestUnconsumedByUserID(ctx, helper.ToPgUUID(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.OTPVerification{
		ID:                 helper.FromPgUUID(row.ID),
		UserID:             userID,
		Email:              "", // not selected by query; use it only for messaging when needed.
		OtpHash:            row.OtpHash,
		ExpiresAt:          row.ExpiresAt.Time,
		ConsumedAt:         nil,
		ResendCount:        int(row.ResendCount),
		VerifyAttemptCount: int(row.VerifyAttemptCount),
	}, nil
}

func (r *OTPVerificationRepo) InvalidateUnconsumedForUser(ctx context.Context, userID uuid.UUID) error {
	return r.s.Queries.OTPInvalidateUnconsumedForUser(ctx, helper.ToPgUUID(userID))
}

func (r *OTPVerificationRepo) ConsumeByID(ctx context.Context, otpID uuid.UUID) error {
	return r.s.Queries.OTPConsumeByID(ctx, helper.ToPgUUID(otpID))
}

func (r *OTPVerificationRepo) IncrementAttempt(ctx context.Context, otpID uuid.UUID) error {
	return r.s.Queries.OTPIncrementAttemptByID(ctx, helper.ToPgUUID(otpID))
}

func (r *OTPVerificationRepo) ConsumeIfAttemptsExceeded(ctx context.Context, otpID uuid.UUID, maxAttempts int) (int64, error) {
	return r.s.Queries.OTPConsumeIfAttemptsExceeded(ctx, db.OTPConsumeIfAttemptsExceededParams{
		ID:                 helper.ToPgUUID(otpID),
		VerifyAttemptCount: int32(maxAttempts),
	})
}

func (r *OTPVerificationRepo) GetLatestIssuedMeta(ctx context.Context, userID uuid.UUID) (*time.Time, int, error) {
	row, err := r.s.Queries.OTPGetLatestIssuedMeta(ctx, helper.ToPgUUID(userID))
	if err != nil {
		return nil, 0, err
	}
	var lastIssuedAt *time.Time
	switch v := row.LastIssuedAt.(type) {
	case nil:
		lastIssuedAt = nil
	case time.Time:
		t := v
		lastIssuedAt = &t
	case pgtype.Timestamptz:
		if v.Valid {
			t := v.Time
			lastIssuedAt = &t
		}
	default:
		// Best-effort: try pgtype cast.
		if ts, ok := v.(pgtype.Timestamptz); ok && ts.Valid {
			t := ts.Time
			lastIssuedAt = &t
		}
	}
	return lastIssuedAt, int(row.LastResendCount), nil
}
