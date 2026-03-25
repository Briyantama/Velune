package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/notification-service/internal/domain"
	"github.com/moon-eye/velune/shared/contracts"
)

type EventConsumer interface {
	Consume(ctx context.Context, handler func(context.Context, contracts.EventEnvelope) error) error
}

type EventPublisher interface {
	Publish(ctx context.Context, envelope contracts.EventEnvelope) error
}

type DeliveryChannel interface {
	Name() string
	Deliver(ctx context.Context, alert domain.OverspendAlert) error
}

type DedupeStore interface {
	SeenOrMark(ctx context.Context, key string, eventID uuid.UUID) (bool, error)
}

type NotificationJob struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Channel     string
	Payload     json.RawMessage
	Status      string
	RetryCount  int
	NextRetryAt time.Time
}

type JobRepository interface {
	Enqueue(ctx context.Context, job *NotificationJob) error
	FetchDue(ctx context.Context, limit int) ([]NotificationJob, error)
	MarkSent(ctx context.Context, id uuid.UUID) error
	MarkRetry(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID) error
}
