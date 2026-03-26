-- name: ProvisioningStateInsertIfMissing :exec
INSERT INTO provisioning_state (id, user_id, account_provisioned_at, version, created_at, updated_at)
VALUES ($1, $2, NULL, $3, now(), now())
ON CONFLICT (user_id) DO NOTHING;

-- name: ProvisioningStateClaimProvisioning :execrows
UPDATE provisioning_state
SET account_provisioned_at = now(),
    updated_at = now(),
    version = version + 1
WHERE user_id = $1
  AND deleted_at IS NULL
  AND account_provisioned_at IS NULL;

