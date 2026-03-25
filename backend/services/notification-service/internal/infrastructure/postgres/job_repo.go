package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/notification-service/internal/repository"
	"github.com/moon-eye/velune/shared/helper"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

type JobRepo struct{ s *Store }

func NewJobRepo(s *Store) *JobRepo { return &JobRepo{s: s} }

func (r *JobRepo) Enqueue(ctx context.Context, job *repository.NotificationJob) error {
	q := db.New(r.s.Pool)
	return q.NotificationJobEnqueue(ctx, db.NotificationJobEnqueueParams{
		ID:      helper.ToPgUUID(job.ID),
		UserID:  helper.ToPgUUID(job.UserID),
		Channel: job.Channel,
		Payload: job.Payload,
	})
}

func (r *JobRepo) FetchDue(ctx context.Context, limit int) ([]repository.NotificationJob, error) {
	q := db.New(r.s.Pool)
	rows, err := q.NotificationJobsFetchDue(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	out := make([]repository.NotificationJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, repository.NotificationJob{
			ID:          helper.FromPgUUID(row.ID),
			UserID:      helper.FromPgUUID(row.UserID),
			Channel:     row.Channel,
			Payload:     row.Payload,
			Status:      row.Status,
			RetryCount:  int(row.RetryCount),
			NextRetryAt: row.NextRetryAt.Time,
		})
	}
	return out, nil
}

func (r *JobRepo) MarkSent(ctx context.Context, id uuid.UUID) error {
	return db.New(r.s.Pool).NotificationJobMarkSent(ctx, helper.ToPgUUID(id))
}

func (r *JobRepo) MarkRetry(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt time.Time) error {
	return db.New(r.s.Pool).NotificationJobMarkRetry(ctx, db.NotificationJobMarkRetryParams{
		ID:          helper.ToPgUUID(id),
		RetryCount:  int32(retryCount),
		NextRetryAt: helper.ToPgTS(nextRetryAt),
	})
}

func (r *JobRepo) MarkFailed(ctx context.Context, id uuid.UUID) error {
	return db.New(r.s.Pool).NotificationJobMarkFailed(ctx, helper.ToPgUUID(id))
}
