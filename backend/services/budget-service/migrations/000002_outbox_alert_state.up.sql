CREATE TABLE IF NOT EXISTS event_outbox (
  id UUID PRIMARY KEY,
  event_type VARCHAR(128) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'pending',
  retry_count INT NOT NULL DEFAULT 0,
  next_retry_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_budget_outbox_status_retry
ON event_outbox(status, next_retry_at, created_at);

CREATE TABLE IF NOT EXISTS budget_alert_state (
  budget_id UUID PRIMARY KEY,
  last_usage_percent DOUBLE PRECISION NOT NULL,
  last_threshold_state VARCHAR(16) NOT NULL,
  last_emitted_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
