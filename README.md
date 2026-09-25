# Finory

This repository is the starting structure for the income and expense tracker described in [AGENTS.md](AGENTS.md).

The `frontend/` app was initialized with `create-next-app@latest` using App Router, TypeScript, Tailwind CSS, and ESLint. Run it from the repository root with:

```sh
make run-frontend
```

The frontend dependencies listed in `AGENTS.md` are installed: shadcn/ui, Lucide React, React Hook Form, Zod with its resolver, TanStack Query, Recharts, Sonner, and date-fns. TanStack Table remains deferred until the transaction table needs advanced behavior.

Each Go backend service starts independently and exposes `GET /health`, returning `{"status":"ok"}`. Run one with `make run-api-gateway`, `make run-auth-service`, `make run-ledger-service`, or `make run-analytics-service`. These explicit Makefile targets, along with `make run-frontend`, can be completed with Tab in shells with Make completion enabled. The default ports are 8080, 8081, 8082, and 8083 respectively; set `PORT` to override a port. Run `make backend-check` to compile and check all four modules.

## Local backend with Docker Compose

Create separate `auth_db` and `ledger_db` databases in Neon, then copy `.env.example` to `.env` at the repository root and replace both example URLs with the corresponding Neon connection strings. Keep `.env` private; Git ignores it. Each service receives only its own database URL. Use TLS enabled URLs (`sslmode=require`).

Run `make compose-up` to build and start API Gateway, Auth Service, Ledger Service, and Analytics Service. Compose waits for their health checks. The gateway is available at `http://localhost:8080/health`. Set `API_GATEWAY_PORT` in `.env` if port 8080 is occupied. Run `make compose-logs` to follow logs and `make compose-down` to stop the stack.

The backend feature folders are placeholders. The services expose only `/health` today; their database URLs and internal service URLs are prepared for later routes but are not used yet. No database schema, authentication flow, or other API routes have been implemented.
