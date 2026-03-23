CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    period_type VARCHAR(32) NOT NULL,
    category_id UUID,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    limit_amount_minor BIGINT NOT NULL CHECK (limit_amount_minor >= 0),
    currency CHAR(3) NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_budgets_user_dates ON budgets(user_id, start_date, end_date) WHERE deleted_at IS NULL;

