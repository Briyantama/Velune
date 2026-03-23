CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(320) NOT NULL,
    password_hash TEXT NOT NULL,
    base_currency CHAR(3) NOT NULL DEFAULT 'USD',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_users_email_active ON users (lower(email)) WHERE deleted_at IS NULL;

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(200) NOT NULL,
    type VARCHAR(32) NOT NULL,
    currency CHAR(3) NOT NULL,
    balance_minor BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_accounts_user ON accounts(user_id) WHERE deleted_at IS NULL;

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(200) NOT NULL,
    parent_id UUID REFERENCES categories(id),
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_categories_user ON categories(user_id) WHERE deleted_at IS NULL;

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    category_id UUID REFERENCES categories(id),
    counterparty_account_id UUID REFERENCES accounts(id),
    amount_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    type VARCHAR(32) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_transactions_user_occurred ON transactions(user_id, occurred_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_account ON transactions(account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_category ON transactions(category_id) WHERE deleted_at IS NULL AND category_id IS NOT NULL;

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(200) NOT NULL,
    period_type VARCHAR(32) NOT NULL,
    category_id UUID REFERENCES categories(id),
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

CREATE TABLE recurring_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    category_id UUID REFERENCES categories(id),
    amount_minor BIGINT NOT NULL CHECK (amount_minor >= 0),
    currency CHAR(3) NOT NULL,
    type VARCHAR(32) NOT NULL,
    frequency VARCHAR(32) NOT NULL,
    next_run_at TIMESTAMPTZ NOT NULL,
    last_run_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    description TEXT NOT NULL DEFAULT '',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_recurring_user_next ON recurring_rules(user_id, next_run_at) WHERE deleted_at IS NULL AND is_active = true;

CREATE TABLE change_events (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    entity_type VARCHAR(64) NOT NULL,
    entity_id UUID NOT NULL,
    operation VARCHAR(16) NOT NULL,
    version BIGINT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_change_events_user_created ON change_events(user_id, created_at DESC);
