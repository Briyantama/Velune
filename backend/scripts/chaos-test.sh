#!/usr/bin/env bash
# Local chaos / resilience checks (simulation flags off in compose by default).
# Requires: Docker, bash. Optional: curl for RabbitMQ management API.
# Usage: `make chaos-test` from [backend/](backend/) or `bash scripts/chaos-test.sh`.
#
# Enable simulation by setting env on services (compose override) or exporting before `docker compose up`, e.g.:
#   SIMULATE_PUBLISH_FAIL_RATE  SIMULATE_BROKER_DOWN  SIMULATE_EMAIL_FAIL_RATE
#   SIMULATE_CONSUMER_PANIC   SIMULATE_DLQ_SNOOP     SIMULATE_SEED

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
INFRA_DIR="${BACKEND_ROOT}/infra"
COMPOSE_FILE="${INFRA_DIR}/docker-compose.yml"

if [[ ! -f "${COMPOSE_FILE}" ]]; then
  echo "compose file not found: ${COMPOSE_FILE}" >&2
  exit 1
fi

cd "${INFRA_DIR}"

echo "==> Starting postgres + rabbitmq + outbox publisher + consumer (split profile)"
docker compose -f docker-compose.yml --profile split up -d postgres rabbitmq transaction-service notification-service

echo "==> Waiting for postgres (15s)"
sleep 15

echo "==> Transaction DB: event_outbox sample"
docker compose -f docker-compose.yml exec -T postgres psql -U postgres -d velune_transaction -c \
  "SELECT id, event_type, status, retry_count, next_retry_at FROM event_outbox ORDER BY created_at DESC LIMIT 10;" || true

echo "==> Notification DB: jobs + dedupe"
docker compose -f docker-compose.yml exec -T postgres psql -U postgres -d velune_notification -c \
  "SELECT id, channel, status, retry_count, next_retry_at FROM notification_jobs ORDER BY created_at DESC LIMIT 10;" || true
docker compose -f docker-compose.yml exec -T postgres psql -U postgres -d velune_notification -c \
  "SELECT idempotency_key, event_id FROM event_dedupe ORDER BY processed_at DESC LIMIT 10;" || true

echo "==> RabbitMQ DLQ queue (optional)"
if command -v curl >/dev/null 2>&1; then
  curl -s -u guest:guest "http://localhost:15672/api/queues/%2F/velune.events.dlq" | head -c 500 || true
  echo ""
else
  echo "curl not installed; skip management API"
fi

echo ""
echo "Simulation env vars (all default off):"
echo "  SIMULATE_BROKER_DOWN SIMULATE_PUBLISH_FAIL_RATE SIMULATE_EMAIL_FAIL_RATE"
echo "  SIMULATE_CONSUMER_PANIC SIMULATE_DLQ_SNOOP SIMULATE_SEED"
echo ""
echo "Tests:  make -C ${BACKEND_ROOT} test-sim"
echo "Dedupe: make -C ${BACKEND_ROOT} test-notification-integration  (needs INTEGRATION_DATABASE_URL)"
