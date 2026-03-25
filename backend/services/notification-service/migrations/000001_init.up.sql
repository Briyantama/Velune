CREATE TABLE IF NOT EXISTS event_dedupe (
  idempotency_key VARCHAR(255) PRIMARY KEY,
  event_id UUID NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS notification_jobs (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  channel VARCHAR(16) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'pending',
  retry_count INT NOT NULL DEFAULT 0,
  next_retry_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notification_jobs_status_retry
ON notification_jobs(status, next_retry_at, created_at);
