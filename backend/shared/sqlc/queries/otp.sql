-- name: OTPInsert :exec
INSERT INTO otp_verifications (
  id, user_id, email, otp_hash, expires_at, consumed_at,
  resend_count, verify_attempt_count,
  version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6,
        $7, $8,
        $9, $10, $11);

-- name: OTPGetLatestUnconsumedByUserID :one
SELECT id, otp_hash, expires_at, verify_attempt_count, resend_count
FROM otp_verifications
WHERE user_id = $1
  AND consumed_at IS NULL
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: OTPConsumeByID :exec
UPDATE otp_verifications
SET consumed_at = now(), updated_at = now(), version = version + 1
WHERE id = $1
  AND consumed_at IS NULL
  AND deleted_at IS NULL;

-- name: OTPConsumeIfAttemptsExceeded :execrows
UPDATE otp_verifications
SET consumed_at = now(), updated_at = now(), version = version + 1
WHERE id = $1
  AND consumed_at IS NULL
  AND deleted_at IS NULL
  AND verify_attempt_count >= $2;

-- name: OTPIncrementAttemptByID :exec
UPDATE otp_verifications
SET verify_attempt_count = verify_attempt_count + 1,
    updated_at = now(),
    version = version + 1
WHERE id = $1
  AND consumed_at IS NULL
  AND deleted_at IS NULL;

-- name: OTPInvalidateUnconsumedForUser :exec
UPDATE otp_verifications
SET consumed_at = now(), updated_at = now(), version = version + 1
WHERE user_id = $1
  AND consumed_at IS NULL
  AND deleted_at IS NULL;

-- name: OTPGetLatestIssuedMeta :one
SELECT
  MAX(created_at) AS last_issued_at,
  COALESCE(MAX(resend_count), 0)::int4                           AS last_resend_count
FROM otp_verifications
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: UserActivateAfterOTP :exec
UPDATE users
SET status = 'active',
    email_verified_at = now(),
    updated_at = now(),
    version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UserGetProvisioningAccountProvisionedAt :one
SELECT account_provisioned_at
FROM provisioning_state
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: ProvisioningStateUpsertMarkAccountProvisioned :exec
INSERT INTO provisioning_state (id, user_id, account_provisioned_at, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id) DO UPDATE
SET account_provisioned_at = COALESCE(provisioning_state.account_provisioned_at, EXCLUDED.account_provisioned_at),
    updated_at = now(),
    version = provisioning_state.version + 1
WHERE provisioning_state.deleted_at IS NULL;

