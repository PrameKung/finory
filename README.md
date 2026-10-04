# Finory

Finory is a personal income and expense tracker with a Next.js frontend and a single Go backend service. The backend is a modular monolith: auth, ledger, and analytics are feature packages in one application, deployed and run together.

## Run locally

The frontend was initialized with Next.js App Router, TypeScript, Tailwind CSS, and ESLint. Its UI dependencies are listed in `frontend/package.json`.

```sh
make run-frontend
```

For the backend, copy `.env.example` to `.env.local`, provide PostgreSQL URLs for `auth_db` and `ledger_db`, a JWT secret of at least 32 bytes, and Google OAuth credentials. The callback URL must exactly match the URL registered with Google:

```text
http://localhost:8080/api/v1/auth/google/callback
```

Run the API directly with `make run-api`, or start the Docker Compose stack with `make compose-up`. The API is available at `http://localhost:8080`; `GET /health` checks the process and `GET /ready` checks the database connection. Use `make compose-logs` and `make compose-down` to inspect and stop Compose.

## Database

The API is one deployed service with two database connections. Auth data stays in `auth_db`; categories, wallets, transactions, and budgets stay in `ledger_db`. Apply each Goose migration set to its matching database before starting the API:

```sh
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$AUTH_DATABASE_URL" goose -dir backend/service/migrations/auth up
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$LEDGER_DATABASE_URL" goose -dir backend/service/migrations/ledger up
```

The auth and ledger databases remain independently owned and use their own migration history. Existing databases from the previous setup can keep their data and migration versions.

Auth and ledger query definitions live under `backend/service/sql/`. Their generated pgx packages are kept under `internal/auth/database` and `internal/ledger/database` respectively. Regenerate them with the matching config from `backend/service`:

```sh
sqlc generate -f sqlc-auth.yaml
sqlc generate -f sqlc-ledger.yaml
```

## API

All public routes are served directly by the one Go process under `/api/v1`. The frontend uses only the configured public API URL. Auth, ledger CRUD, and analytics handlers call their application services in-process; user-owned ledger queries remain scoped by the authenticated user ID.

The interactive Scalar API reference is at `http://localhost:8080/docs`, backed by `/openapi.json`.

## Commands

- `make run-api` starts the Go backend.
- `make run-frontend` starts Next.js.
- `make compose-up`, `make compose-logs`, and `make compose-down` manage local containers.
- `make backend-check` runs Go package tests.
- `make frontend-path-test` runs the Docker-backed transaction path integration test (Docker required).
