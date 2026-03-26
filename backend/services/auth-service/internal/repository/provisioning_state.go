package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/auth-service/internal/domain"
)

type ProvisioningStateRepository interface {
	GetAccountProvisionedAt(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	MarkAccountProvisionedAt(ctx context.Context, userID uuid.UUID, at time.Time) error
}

// Provisioning state is auth-owned, but it is updated based on transaction-service events.
var _ = domain.ProvisioningState{}
