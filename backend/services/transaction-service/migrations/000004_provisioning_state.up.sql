-- Provisioning marker for auth-service first-login default account provisioning.
CREATE TABLE IF NOT EXISTS provisioning_state (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    account_provisioned_at TIMESTAMPTZ,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_provisioning_state_user_provisioned_at
  ON provisioning_state(user_id, account_provisioned_at)
  WHERE deleted_at IS NULL;

