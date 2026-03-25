# Run from repository root: backend/  (same folder as go.work)
#   make help
#   make deps           # Postgres + RabbitMQ + Redis (detached)
#   make run-auth       # in another terminal, after deps

.PHONY: help deps deps-down deps-logs \
	compose-legacy compose-split compose-down \
	work-sync build tidy test fmt \
	run-auth run-transaction run-budget run-report run-notification \
	run-category run-gateway run-legacy run-admin \
	chaos-test test-sim test-notification-integration

ROOT := $(CURDIR)
COMPOSE := docker compose -f $(ROOT)/infra/docker-compose.yml
# Local Postgres/RabbitMQ ports (match infra/docker-compose.yml)
PG_HOST ?= 127.0.0.1
PG_PORT ?= 5432
PG_USER ?= postgres
PG_PASS ?= postgres
PG_BASE := postgres://$(PG_USER):$(PG_PASS)@$(PG_HOST):$(PG_PORT)
JWT_SECRET ?= dev-only-change-me
ADMIN_API_KEY ?= dev-admin-api-key
ADMIN_INTERNAL_KEY ?= dev-admin-internal-key
BROKER_URL ?= amqp://guest:guest@$(PG_HOST):5672/

## help: print targets
help:
	@echo "Velune backend — local development"
	@echo ""
	@echo "Infrastructure (Docker):"
	@echo "  make deps           Postgres + RabbitMQ + Redis (init script creates DBs)"
	@echo "  make deps-down      stop infra containers"
	@echo "  make deps-logs      follow infra logs"
	@echo "  make compose-legacy legacy-api + postgres + redis (no RabbitMQ in stack)"
	@echo "  make compose-split  full split stack + gateway (builds images)"
	@echo "  make compose-down   docker compose down"
	@echo ""
	@echo "Go workspace:"
	@echo "  make work-sync      go work sync"
	@echo "  make build          go build all workspace modules"
	@echo "  make tidy           go mod tidy in shared + each service"
	@echo "  make test           go test ./... per module (best effort)"
	@echo "  make fmt            go fmt ./... under shared + services"
	@echo ""
	@echo "Run services natively (need: make deps; set PG_HOST if not localhost):"
	@echo "  make run-auth | run-transaction | run-budget | run-report"
	@echo "  make run-notification | run-category | run-gateway | run-legacy | run-admin"
	@echo ""
	@echo "Other:"
	@echo "  make chaos-test | test-sim | test-notification-integration"

## deps: start Postgres (multi-DB init), RabbitMQ management UI :15672, Redis
deps:
	$(COMPOSE) up -d rabbitmq redis

## deps-down: stop infra containers
deps-down:
	$(COMPOSE) down

## deps-logs: follow infra
deps-logs:
	$(COMPOSE) logs -f rabbitmq redis

## compose-legacy: legacy API + DB + redis (see README)
compose-legacy:
	$(COMPOSE) up legacy-api redis

## compose-split: all microservices + gateway (Docker build)
compose-split:
	$(COMPOSE) --profile split up --build

## compose-down: tear down compose project
compose-down:
	$(COMPOSE) down

work-sync:
	cd "$(ROOT)" && go work sync

## build: compile all modules in go.work
build: work-sync
	cd "$(ROOT)/shared" && go build ./...
	cd "$(ROOT)/services/admin-service" && go build ./...
	cd "$(ROOT)/services/api-gateway" && go build ./...
	cd "$(ROOT)/services/auth-service" && go build ./...
	cd "$(ROOT)/services/budget-service" && go build ./...
	cd "$(ROOT)/services/category-service" && go build ./...
	cd "$(ROOT)/services/legacy-api" && go build ./...
	cd "$(ROOT)/services/notification-service" && go build ./...
	cd "$(ROOT)/services/report-service" && go build ./...
	cd "$(ROOT)/services/transaction-service" && go build ./...

tidy:
	cd "$(ROOT)/shared" && go mod tidy
	cd "$(ROOT)/services/admin-service" && go mod tidy
	cd "$(ROOT)/services/api-gateway" && go mod tidy
	cd "$(ROOT)/services/auth-service" && go mod tidy
	cd "$(ROOT)/services/budget-service" && go mod tidy
	cd "$(ROOT)/services/category-service" && go mod tidy
	cd "$(ROOT)/services/legacy-api" && go mod tidy
	cd "$(ROOT)/services/notification-service" && go mod tidy
	cd "$(ROOT)/services/report-service" && go mod tidy
	cd "$(ROOT)/services/transaction-service" && go mod tidy

test:
	cd "$(ROOT)/shared" && go test ./...
	cd "$(ROOT)/services/admin-service" && go test ./...
	cd "$(ROOT)/services/api-gateway" && go test ./...
	cd "$(ROOT)/services/auth-service" && go test ./...
	cd "$(ROOT)/services/budget-service" && go test ./...
	cd "$(ROOT)/services/category-service" && go test ./...
	cd "$(ROOT)/services/legacy-api" && go test ./...
	cd "$(ROOT)/services/notification-service" && go test ./...
	cd "$(ROOT)/services/report-service" && go test ./...
	cd "$(ROOT)/services/transaction-service" && go test ./...

fmt:
	cd "$(ROOT)/shared" && go fmt ./...
	cd "$(ROOT)/services/admin-service" && go fmt ./...
	cd "$(ROOT)/services/api-gateway" && go fmt ./...
	cd "$(ROOT)/services/auth-service" && go fmt ./...
	cd "$(ROOT)/services/budget-service" && go fmt ./...
	cd "$(ROOT)/services/category-service" && go fmt ./...
	cd "$(ROOT)/services/legacy-api" && go fmt ./...
	cd "$(ROOT)/services/notification-service" && go fmt ./...
	cd "$(ROOT)/services/report-service" && go fmt ./...
	cd "$(ROOT)/services/transaction-service" && go fmt ./...

# --- Native runs (cwd = backend root; migrations paths point into each service) ---

run-auth:
	cd "$(ROOT)" && \
	DATABASE_URL="$(PG_BASE)/velune_auth?sslmode=disable" \
	JWT_SECRET="$(JWT_SECRET)" \
	MIGRATIONS_PATH=file://services/auth-service/migrations \
	LOG_FORMAT=dev \
	go run ./services/auth-service/cmd/api

run-transaction:
	cd "$(ROOT)" && \
	DATABASE_URL="$(PG_BASE)/velune_transaction?sslmode=disable" \
	JWT_SECRET="$(JWT_SECRET)" \
	MIGRATIONS_PATH=file://services/transaction-service/migrations \
	BROKER_URL="$(BROKER_URL)" \
	BROKER_EXCHANGE=velune.events \
	BROKER_ROUTING_KEY=notification.overspend.requested \
	BROKER_DLX=velune.events.dlx \
	BROKER_DLQ=velune.events.dlq \
	BROKER_DLQ_ROUTING_KEY=velune.dlq \
	ADMIN_INTERNAL_KEY="$(ADMIN_INTERNAL_KEY)" \
	LOG_FORMAT=dev \
	go run ./services/transaction-service/cmd/api

run-budget:
	cd "$(ROOT)" && \
	DATABASE_URL="$(PG_BASE)/velune_budget?sslmode=disable" \
	JWT_SECRET="$(JWT_SECRET)" \
	MIGRATIONS_PATH=file://services/budget-service/migrations \
	TRANSACTION_SERVICE_URL=http://$(PG_HOST):8082 \
	BROKER_URL="$(BROKER_URL)" \
	BROKER_EXCHANGE=velune.events \
	BROKER_ROUTING_KEY=notification.overspend.requested \
	BROKER_DLX=velune.events.dlx \
	BROKER_DLQ=velune.events.dlq \
	BROKER_DLQ_ROUTING_KEY=velune.dlq \
	LOG_FORMAT=dev \
	go run ./services/budget-service/cmd/api

run-report:
	cd "$(ROOT)" && \
	JWT_SECRET="$(JWT_SECRET)" \
	TRANSACTION_SERVICE_URL=http://$(PG_HOST):8082 \
	LOG_FORMAT=dev \
	go run ./services/report-service/cmd/api

run-notification:
	cd "$(ROOT)" && \
	DATABASE_URL="$(PG_BASE)/velune_notification?sslmode=disable" \
	JWT_SECRET="$(JWT_SECRET)" \
	MIGRATIONS_PATH=file://services/notification-service/migrations \
	BROKER_URL="$(BROKER_URL)" \
	BROKER_EXCHANGE=velune.events \
	BROKER_QUEUE=notification.overspend \
	BROKER_ROUTING_KEY=notification.overspend.requested \
	BROKER_DLX=velune.events.dlx \
	BROKER_DLQ=velune.events.dlq \
	BROKER_DLQ_ROUTING_KEY=velune.dlq \
	EMAIL_FROM=noreply@velune.local \
	LOG_FORMAT=dev \
	go run ./services/notification-service/cmd/api

run-category:
	cd "$(ROOT)" && \
	LOG_FORMAT=dev \
	go run ./services/category-service/cmd/server

run-gateway:
	cd "$(ROOT)" && \
	AUTH_SERVICE_URL=http://$(PG_HOST):8081 \
	TRANSACTION_SERVICE_URL=http://$(PG_HOST):8082 \
	CATEGORY_SERVICE_URL=http://$(PG_HOST):8083 \
	BUDGET_SERVICE_URL=http://$(PG_HOST):8084 \
	REPORT_SERVICE_URL=http://$(PG_HOST):8085 \
	NOTIFICATION_SERVICE_URL=http://$(PG_HOST):8086 \
	LEGACY_API_URL=http://$(PG_HOST):8090 \
	LOG_FORMAT=dev \
	go run ./services/api-gateway/cmd/server

run-legacy:
	cd "$(ROOT)" && \
	HTTP_HOST=0.0.0.0 \
	HTTP_PORT=8090 \
	JWT_SECRET="$(JWT_SECRET)" \
	LOG_FORMAT=dev \
	go run ./services/legacy-api/cmd/api

run-admin:
	cd "$(ROOT)" && \
	HTTP_PORT=8099 \
	ADMIN_API_KEY="$(ADMIN_API_KEY)" \
	ADMIN_INTERNAL_KEY="$(ADMIN_INTERNAL_KEY)" \
	TRANSACTION_DATABASE_URL="$(PG_BASE)/velune_transaction?sslmode=disable" \
	BUDGET_DATABASE_URL="$(PG_BASE)/velune_budget?sslmode=disable" \
	NOTIFICATION_DATABASE_URL="$(PG_BASE)/velune_notification?sslmode=disable" \
	BROKER_URL="$(BROKER_URL)" \
	BROKER_EXCHANGE=velune.events \
	BROKER_ROUTING_KEY=notification.overspend.requested \
	BROKER_DLX=velune.events.dlx \
	BROKER_DLQ=velune.events.dlq \
	BROKER_DLQ_ROUTING_KEY=velune.dlq \
	AUTH_SERVICE_URL=http://$(PG_HOST):8081 \
	TRANSACTION_SERVICE_URL=http://$(PG_HOST):8082 \
	CATEGORY_SERVICE_URL=http://$(PG_HOST):8083 \
	BUDGET_SERVICE_URL=http://$(PG_HOST):8084 \
	REPORT_SERVICE_URL=http://$(PG_HOST):8085 \
	NOTIFICATION_SERVICE_URL=http://$(PG_HOST):8086 \
	LOG_FORMAT=dev \
	go run ./services/admin-service/cmd/api

chaos-test:
	cd "$(ROOT)" && bash scripts/chaos-test.sh

test-sim:
	cd "$(ROOT)/shared/sim" && go test -count=1 ./...

test-notification-integration:
	cd "$(ROOT)/services/notification-service" && go test -tags=integration -count=1 ./internal/infrastructure/postgres/...
