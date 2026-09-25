# Finory

This repository is the starting structure for the income and expense tracker described in [AGENTS.md](AGENTS.md).

The `frontend/` app was initialized with `create-next-app@latest` using App Router, TypeScript, Tailwind CSS, and ESLint. Run it from the repository root with:

```sh
make run-frontend
```

The frontend dependencies listed in `AGENTS.md` are installed: shadcn/ui, Lucide React, React Hook Form, Zod with its resolver, TanStack Query, Recharts, Sonner, and date-fns. TanStack Table remains deferred until the transaction table needs advanced behavior.

Each Go backend service uses Echo, starts independently, and exposes `GET /health`, returning `{"status":"ok"}`. Run one with `make run-api-gateway`, `make run-auth-service`, `make run-ledger-service`, or `make run-analytics-service`. Auth and Ledger require `DATABASE_URL` when run outside Compose. These explicit Makefile targets, along with `make run-frontend`, can be completed with Tab in shells with Make completion enabled. The default ports are 8080, 8081, 8082, and 8083 respectively; set `PORT` to override a port. Run `make backend-check` to compile and check all four modules.

## Local backend with Docker Compose

Create separate `auth_db` and `ledger_db` databases in Neon, then copy `.env.example` to `.env.local` at the repository root and replace both example URLs with the corresponding Neon connection strings. Keep `.env.local` private; Git ignores it. Each service receives only its own database URL. Use TLS enabled URLs (`sslmode=require`).

Run `make compose-up` to build and start API Gateway, Auth Service, Ledger Service, and Analytics Service under the `finory` Compose project name. Auth and Ledger confirm they connected to `auth_db` and `ledger_db` respectively before listening. Their `GET /ready` endpoints check their database connections on demand. Compose uses `/health` for routine checks so it does not repeatedly query Neon. The gateway is available at `http://localhost:8080/health`. Set `API_GATEWAY_PORT` in `.env.local` if port 8080 is occupied. Run `make compose-logs` to follow logs and `make compose-down` to stop the stack. To inspect readiness from inside the containers, run `docker compose --env-file .env.local -f deployments/docker/compose.yaml exec auth-service wget -qO- http://localhost:8081/ready` and the equivalent command with `ledger-service` and port `8082`.

The backend feature folders are placeholders. Auth and Ledger connect to Neon, but Ledger has no database schema yet and neither service has domain routes. The Analytics Service does not use a database. The gateway forwards `/api/v1/auth` and `/api/v1/auth/*` to the Auth Service as `/auth` and `/auth/*`, except `GET /api/v1/auth/health`, which forwards to the Auth Service's existing `/health` endpoint. It forwards `/api/v1/transactions`, `/api/v1/categories`, `/api/v1/wallets`, and `/api/v1/budgets` and their child paths to the Ledger Service, removing only `/api/v1` from each path. It forwards `/api/v1/analytics` and its child paths to the Analytics Service as `/analytics` and its child paths, except `GET /api/v1/analytics/health`, which forwards to the Analytics Service's existing `/health` endpoint. Auth, Ledger, and Analytics domain handlers for these paths are still pending.

The Auth Service's Goose migration in `backend/services/auth-service/migrations/` creates the `users` table in `auth_db`. It stores a UUID user ID, a unique Google `sub` identifier, email, optional display name and avatar URL, and UTC-compatible timestamps. Email is not an account key because it can change. There is no password column. Apply this migration with Goose using the Auth Service database URL only:

```sh
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$AUTH_DATABASE_URL" goose -dir backend/services/auth-service/migrations up
```

Each Go service loads and validates its own environment variables at startup. `PORT` defaults to `8080` for the gateway, `8081` for Auth, `8082` for Ledger, and `8083` for Analytics. Auth and Ledger require `DATABASE_URL`; Compose supplies each from its matching root `.env.local` variable. For standalone runs, export `DATABASE_URL` in the service process environment.

The gateway accepts `AUTH_SERVICE_URL`, `LEDGER_SERVICE_URL`, and `ANALYTICS_SERVICE_URL`, defaulting to `http://localhost:8081`, `http://localhost:8082`, and `http://localhost:8083`. Analytics accepts `LEDGER_SERVICE_URL`, defaulting to `http://localhost:8082`. Compose sets these to the internal service names. The gateway uses all three service URLs to forward their routes.

The gateway handles CORS for all routes. It allows `http://localhost:3000` and `http://127.0.0.1:3000` by default. Set `CORS_ALLOWED_ORIGINS` to a comma-separated list of frontend origins for other environments, such as `https://app.example.com`. Preflight requests are answered by the gateway; backend services do not need their own CORS configuration. To check locally, run `curl -i -X OPTIONS http://localhost:8080/api/v1/auth/login -H 'Origin: http://localhost:3000' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: content-type,authorization'`.

Every gateway response includes `X-Request-ID`. The gateway preserves a supplied ID or generates one when absent, then forwards the same ID to the internal service. Use `curl -i -H 'X-Request-ID: manual-check-1' http://localhost:8080/api/v1/auth/health` to see it in the response.

The gateway requires `JWT_ACCESS_SECRET` (at least 32 bytes) to verify HS256 access tokens with a UUID `sub` and an `exp` claim. Set it in `.env.local` for Compose or export it for `make run-api-gateway`; the Auth Service must use the same access-token secret when token issuance is implemented. Registration, login, refresh, and health checks remain public. Other Auth paths, all Ledger paths, and Analytics paths other than its health check require `Authorization: Bearer <access-token>`. The gateway returns `401` for missing, malformed, expired, or invalid tokens. The Auth Service does not issue access tokens yet, so protected routes cannot be used end to end until that flow is implemented.
