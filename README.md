# Velune

Expense tracker, money manager, and budget planner—built for clean architecture, maintainability, and production-grade operations.

## Architecture

- **Clean Architecture**: dependency direction flows **Domain → Application (use cases) → Ports / interfaces → Infrastructure**; HTTP handlers and workers live in **Delivery** and must not contain business rules.
- **Domain language**: Transaction, Account, Category, Budget, Recurring, Report (see `.cursor/rules/app-expense-tracker.mdc` for full product and layering rules).
- **Backend**: Go services and conventions are described in `backend/.cursor/rules/golang-backend.mdc` (stack, boundaries, observability expectations).

## Repositories

| Area    | Location   | Notes                                      |
|---------|------------|--------------------------------------------|
| Backend | `backend/` | API, workers, persistence, integrations    |
| App rules | `.cursor/rules/` | Shared Cursor rules for this product |

## Setup

1. Clone the repository and open it in your editor.
2. Configure the backend using **environment variables** only (no secrets in code). Each service’s README or `env.example` (when present) documents required variables.
3. Follow language-specific setup in `backend/README.md` once the backend layout is initialized (`go mod`, database migrations, etc.).

## API contract

When HTTP APIs are exposed, maintain a versioned **OpenAPI** (or equivalent) specification in the backend repository and update it whenever endpoints or schemas change.

## Contributing

- Branch naming: `feat/<scope>`, `fix/<scope>` (for example `feat/transaction-create`).
- Submit changes via **Pull Request** with review before merge.
- Keep documentation and API contracts in sync with code changes.
