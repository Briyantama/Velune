package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

// ProvisioningService provisions a user's default financial accounts on first login.
// It is designed to be idempotent under at-least-once delivery of provisioning requests.
type ProvisioningService struct {
	Pool *pgxpool.Pool
}

func NewProvisioningService(pool *pgxpool.Pool) *ProvisioningService {
	return &ProvisioningService{Pool: pool}
}

// ProvisionDefaultAccountsIfNeeded creates the default "Main Wallet" account exactly once
// per user and emits a completion event via the transactional outbox.
func (s *ProvisioningService) ProvisionDefaultAccountsIfNeeded(ctx context.Context, payload contracts.UserFirstLoginProvisionRequested) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := db.New(tx)

	// Ensure marker row exists.
	if err := qtx.ProvisioningStateInsertIfMissing(ctx, db.ProvisioningStateInsertIfMissingParams{
		ID:      helper.ToPgUUID(uuid.New()),
		UserID:  helper.ToPgUUID(payload.UserID),
		Version: 1,
	}); err != nil {
		return err
	}

	// Claim the provisioning responsibility atomically.
	rows, err := qtx.ProvisioningStateClaimProvisioning(ctx, helper.ToPgUUID(payload.UserID))
	if err != nil {
		return err
	}
	if rows == 0 {
		// Already provisioned (or claimed by another concurrent handler).
		return tx.Commit(ctx)
	}

	now := time.Now().UTC()
	// Create default account(s) (only the claim winner reaches this point).
	if err := qtx.AccountCreate(ctx, db.AccountCreateParams{
		ID:           helper.ToPgUUID(uuid.New()),
		UserID:       helper.ToPgUUID(payload.UserID),
		Name:         "Main Wallet",
		Type:         "wallet",
		Currency:     payload.BaseCurrency,
		BalanceMinor: 0,
		Version:      1,
		CreatedAt:    helper.ToPgTS(now),
		UpdatedAt:    helper.ToPgTS(now),
	}); err != nil {
		return err
	}

	completedPayload := contracts.UserFirstLoginProvisionCompleted{
		UserID:     payload.UserID,
		OccurredAt: now,
	}
	completedPayloadBytes, err := helper.ToJSONMarshal(completedPayload)
	if err != nil {
		return err
	}

	env := contracts.EventEnvelope{
		EventID:     uuid.New(),
		EventType:   contracts.EventUserFirstLoginProvisionCompleted,
		Source:      "transaction-service",
		OccurredAt:  now,
		UserID:      &payload.UserID,
		Idempotency: "provision_completed:" + payload.UserID.String(),
		Payload:     completedPayloadBytes,
	}

	envBytes, err := helper.ToJSONMarshal(env)
	if err != nil {
		return err
	}

	if err := qtx.OutboxInsert(ctx, db.OutboxInsertParams{
		ID:        helper.ToPgUUID(env.EventID),
		EventType: env.EventType,
		Payload:   envBytes,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Compile-time guard: pgx.Tx is what sqlc expects for q.WithTx.
var _ pgx.Tx
