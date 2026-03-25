CREATE TABLE budgets (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  name VARCHAR(200) NOT NULL,
  period_type VARCHAR(32) NOT NULL,
  category_id UUID,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  limit_amount_minor BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,
  version BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE budget_alert_state (
  budget_id UUID PRIMARY KEY,
  last_usage_percent DOUBLE PRECISION NOT NULL,
  last_threshold_state VARCHAR(16) NOT NULL,
  last_emitted_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
