package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/notification-service/internal/repository"
)

type JobRepo struct{ s *Store }

func NewJobRepo(s *Store) *JobRepo { return &JobRepo{s: s} }

func (r *JobRepo) Enqueue(ctx context.Context, job *repository.NotificationJob) error {
	_, err := r.s.Pool.Exec(ctx, `
		INSERT INTO notification_jobs (id, user_id, channel, payload, status, retry_count, next_retry_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', 0, now(), now(), now())
	`, job.ID, job.UserID, job.Channel, job.Payload)
	return err
}

func (r *JobRepo) FetchDue(ctx context.Context, limit int) ([]repository.NotificationJob, error) {
	rows, err := r.s.Pool.Query(ctx, `
		SELECT id, user_id, channel, payload, status, retry_count, next_retry_at
		FROM notification_jobs
		WHERE status IN ('pending','processing') AND next_retry_at <= now()
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]repository.NotificationJob, 0, limit)
	for rows.Next() {
		var j repository.NotificationJob
		if err := rows.Scan(&j.ID, &j.UserID, &j.Channel, &j.Payload, &j.Status, &j.RetryCount, &j.NextRetryAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *JobRepo) MarkSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.s.Pool.Exec(ctx, `UPDATE notification_jobs SET status='sent', updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *JobRepo) MarkRetry(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt time.Time) error {
	_, err := r.s.Pool.Exec(ctx, `
		UPDATE notification_jobs
		SET status='processing', retry_count=$2, next_retry_at=$3, updated_at=now()
		WHERE id=$1
	`, id, retryCount, nextRetryAt)
	return err
}

func (r *JobRepo) MarkFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.s.Pool.Exec(ctx, `UPDATE notification_jobs SET status='failed', updated_at=now() WHERE id=$1`, id)
	return err
}
