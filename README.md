# Finory

This repository is the starting structure for the income and expense tracker described in [AGENTS.md](AGENTS.md).

The `frontend/` app was initialized with `create-next-app@latest` using App Router, TypeScript, Tailwind CSS, and ESLint. Run it from the repository root with:

```sh
make run-frontend
```

The frontend dependencies listed in `AGENTS.md` are installed: shadcn/ui, Lucide React, React Hook Form, Zod with its resolver, TanStack Query, Recharts, Sonner, and date-fns. TanStack Table remains deferred until the transaction table needs advanced behavior.

Each Go backend service uses Echo, starts independently, and exposes `GET /health`, returning `{"status":"ok"}`. Run one with `make run-api-gateway`, `make run-auth-service`, `make run-ledger-service`, or `make run-analytics-service`. Auth and Ledger require `DATABASE_URL` when run outside Compose. These explicit Makefile targets, along with `make run-frontend`, can be completed with Tab in shells with Make completion enabled. The default ports are 8080, 8081, 8082, and 8083 respectively; set `PORT` to override a port. Run `make backend-check` to compile and check all four modules.

## Local backend with Docker Compose

Create separate `auth_db` and `ledger_db` databases in Neon, then copy `.env.example` to `.env.local` at the repository root and replace both example URLs with the corresponding Neon connection strings. Keep `.env.local` private; Git ignores it. Each service receives only its own database URL. Use TLS enabled URLs (`sslmode=require`). Create a Google OAuth web client and register the exact `GOOGLE_REDIRECT_URL` as an authorized redirect URI. Set `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, `APP_REDIRECT_URL`, and the same `JWT_ACCESS_SECRET` for the Auth Service and gateway. For local development, the callback URL is `http://localhost:8080/api/v1/auth/google/callback`.

Run `make compose-up` to build and start API Gateway, Auth Service, Ledger Service, and Analytics Service under the `finory` Compose project name. Auth and Ledger confirm they connected to `auth_db` and `ledger_db` respectively before listening. Their `GET /ready` endpoints check their database connections on demand. Compose uses `/health` for routine checks so it does not repeatedly query Neon. The gateway is available at `http://localhost:8080/health`. Set `API_GATEWAY_PORT` in `.env.local` if port 8080 is occupied. Run `make compose-logs` to follow logs and `make compose-down` to stop the stack. To inspect readiness from inside the containers, run `docker compose --env-file .env.local -f deployments/docker/compose.yaml exec auth-service wget -qO- http://localhost:8081/ready` and the equivalent command with `ledger-service` and port `8082`.

Auth and Ledger connect to Neon. The Ledger schema contains user-owned categories and wallets, and the category routes are implemented. The Analytics Service does not use a database. The gateway forwards `/api/v1/auth/*` to the Auth Service as `/auth/*`, except `GET /api/v1/auth/health`, which forwards to `/health`. It forwards Ledger and Analytics paths to their respective services after removing `/api/v1`.

The Auth Service's Goose migrations in `backend/services/auth-service/migrations/` create the `users` and `refresh_sessions` tables in `auth_db`. Users have a UUID ID and unique Google `sub`; email is not an account key because it can change. Refresh sessions store only a hash of each opaque token. There is no password column. Apply these migrations with Goose using the Auth Service database URL only:

```sh
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$AUTH_DATABASE_URL" goose -dir backend/services/auth-service/migrations up
```

The Auth Service's user upsert is defined in `backend/services/auth-service/sql/queries/users.sql`. Regenerate its pgx code from the migration and query with `sqlc generate` in `backend/services/auth-service`.

The Ledger Service's Goose migrations in `backend/services/ledger-service/migrations/` create the `categories` and `wallets` tables in `ledger_db`. Apply them with the Ledger database URL only:

```sh
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$LEDGER_DATABASE_URL" goose -dir backend/services/ledger-service/migrations up
```

Authenticated category endpoints are `GET /api/v1/categories`, `POST /api/v1/categories`, `PATCH /api/v1/categories/:id`, and `DELETE /api/v1/categories/:id`. Creating a category requires `name` and a `type` of `income` or `expense`; `icon` and `color` are optional. Listing categories initializes the standard defaults for a new user. Default categories are read only, while custom categories can be updated and deleted. Every database mutation includes the authenticated user ID so another user's category is returned as not found. Regenerate Ledger's pgx code with `sqlc generate` in `backend/services/ledger-service` after changing its SQL queries.

Each Go service loads and validates its own environment variables at startup. `PORT` defaults to `8080` for the gateway, `8081` for Auth, `8082` for Ledger, and `8083` for Analytics. Auth and Ledger require `DATABASE_URL`; Compose supplies each from its matching root `.env.local` variable. For standalone runs, export `DATABASE_URL` in the service process environment.

Open `http://localhost:8080/api/v1/auth/google` in a browser to start sign-in. The Auth Service checks OAuth state, PKCE, and the Google ID token's signature, audience, issuer, and nonce before creating or updating a user by Google `sub`. It rejects missing, duplicate, or malformed callback data. It sets a 15-minute HTTP-only application JWT cookie and a rolling 30-day HTTP-only refresh cookie, then redirects to `APP_REDIRECT_URL`. The gateway accepts the access cookie or a bearer JWT for protected requests. `GET /api/v1/auth/me` returns the authenticated user's `id`, `email`, `displayName`, and `avatarUrl` from `auth_db`; pass the access cookie with `credentials: "include"` or use a bearer JWT. Browser requests that change data and use a cookie must carry an allowed `Origin` header. Call `POST /api/v1/auth/refresh` with `credentials: "include"` to rotate the refresh token and obtain a new access cookie; an invalid or expired refresh token returns `401`. `POST /api/v1/auth/logout` deletes the refresh session and clears the application and OAuth cookies. JWTs already copied elsewhere remain valid until their 15-minute expiry. No email/password routes exist.

The gateway accepts `AUTH_SERVICE_URL`, `LEDGER_SERVICE_URL`, and `ANALYTICS_SERVICE_URL`, defaulting to `http://localhost:8081`, `http://localhost:8082`, and `http://localhost:8083`. Analytics accepts `LEDGER_SERVICE_URL`, defaulting to `http://localhost:8082`. Compose sets these to the internal service names. The gateway uses all three service URLs to forward their routes.

The gateway handles CORS for all routes and allows credentials from `http://localhost:3000` and `http://127.0.0.1:3000` by default. Set `CORS_ALLOWED_ORIGINS` to a comma-separated list of frontend origins for other environments, such as `https://app.example.com`. Browser API calls using the cookie need `credentials: "include"`. Preflight requests are answered by the gateway; backend services do not need their own CORS configuration.

Every gateway response includes `X-Request-ID`. The gateway preserves a supplied ID or generates one when absent, then forwards the same ID to the internal service. Use `curl -i -H 'X-Request-ID: manual-check-1' http://localhost:8080/api/v1/auth/health` to see it in the response.

The gateway requires `JWT_ACCESS_SECRET` (at least 32 bytes) to verify HS256 access tokens with a UUID `sub` and an `exp` claim. Google authorization, callback, and health checks are public. Other Auth paths, all Ledger paths, and Analytics paths other than its health check require a valid application JWT. The gateway returns `401` for missing, malformed, expired, or invalid tokens.
