.PHONY: chaos-test test-sim test-notification-integration

# Run docker compose snippet and print outbox / notification / DLQ-oriented state.
chaos-test:
	bash scripts/chaos-test.sh

test-sim:
	cd shared/sim && go test -count=1 ./...

# Requires INTEGRATION_DATABASE_URL and Postgres with notification migrations applied.
test-notification-integration:
	cd services/notification-service && go test -tags=integration -count=1 ./internal/infrastructure/postgres/...
