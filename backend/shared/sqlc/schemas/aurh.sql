CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(320) NOT NULL,
  password_hash TEXT NOT NULL,
  base_currency CHAR(3) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  email_verified_at TIMESTAMPTZ,
  display_name VARCHAR(200),
  version BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  version BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ
);

-- Stores hashed OTPs for email verification.
-- We keep otp_hash instead of plaintext and track attempts and consumption.
CREATE TABLE otp_verifications (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  email VARCHAR(320) NOT NULL,
  otp_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  resend_count INT NOT NULL DEFAULT 0,
  verify_attempt_count INT NOT NULL DEFAULT 0,
  version BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT uq_otp_verifications_active UNIQUE (id)
);

-- Single-row marker per user for provisioning completion.
CREATE TABLE provisioning_state (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL UNIQUE,
  account_provisioned_at TIMESTAMPTZ,
  version BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ
);
