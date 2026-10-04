# AGENT.md

## Project Overview

This project is a personal income-expense tracking web application that should be designed to grow into a small SaaS product.

Primary goals:

- Record income and expense transactions
- Manage categories and wallets
- Show monthly summaries and a basic financial dashboard
- Keep every user's data isolated
- Support future SaaS expansion
- Keep the architecture maintainable and easy to extend
- Use the project to learn modular Go architecture, PostgreSQL, Docker, and future deployment options

The MVP should remain small and practical. Avoid adding infrastructure or abstractions that are not yet required.

---

## Repository Structure

Use a monorepo.

```text
expense-tracker/
├── frontend/
├── backend/
│   └── service/
│       ├── cmd/api/
│       ├── internal/
│       └── migrations/
├── deployments/
│   ├── docker/
│   └── kubernetes/
├── docs/
├── Makefile
├── .gitignore
└── README.md
```

Do not introduce Turborepo or Nx unless the repository later becomes complex enough to justify them.

---

# Frontend

## Stack

Use:

- Next.js
- App Router
- TypeScript
- Tailwind CSS
- shadcn/ui
- Lucide React
- React Hook Form
- Zod
- `@hookform/resolvers`
- TanStack Query
- Recharts
- Sonner
- date-fns

Add TanStack Table only when the transaction table needs advanced filtering, sorting, or pagination.

Do not use Redux unless a real global-state problem appears.

---

## Frontend Structure

```text
frontend/
├── public/
├── src/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── globals.css
│   │   │
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   │   ├── layout.tsx
│   │   │   │   └── page.tsx
│   │   │   └── register/
│   │   │       └── page.tsx
│   │   │
│   │   └── (app)/
│   │       ├── layout.tsx
│   │       ├── dashboard/
│   │       ├── transactions/
│   │       ├── categories/
│   │       ├── wallets/
│   │       └── settings/
│   │
│   ├── components/
│   │   ├── ui/
│   │   ├── layout/
│   │   └── shared/
│   │
│   ├── features/
│   │   ├── auth/
│   │   ├── transactions/
│   │   ├── categories/
│   │   ├── wallets/
│   │   └── dashboard/
│   │
│   ├── lib/
│   │   ├── api/
│   │   ├── env.ts
│   │   ├── query-client.ts
│   │   ├── constants.ts
│   │   └── utils.ts
│   │
│   ├── providers/
│   ├── hooks/
│   └── types/
│
├── .env.local
├── .env.example
├── components.json
├── next.config.ts
├── package.json
├── tsconfig.json
└── eslint.config.mjs
```

---

## Frontend Architecture Rules

`app/` is responsible for routing and layouts.

`features/` is responsible for business-feature code.

Example:

```text
features/
└── transactions/
    ├── api/
    ├── components/
    ├── hooks/
    ├── schemas/
    └── types/
```

Keep `page.tsx` files thin. A page should mainly compose feature components.

Do not place business logic directly inside route files unless it is trivial.

The login page may use a special two-column layout.

The register page should use a normal full-page layout.

The frontend communicates with the single Go API service through one public API base URL.

Example:

```text
Next.js -> Go API
```

Use one public API base URL.

```env
NEXT_PUBLIC_API_URL=https://api.example.com
```

---

# Backend

## Architecture

The backend is one deployable Go service built as a modular monolith. Auth, ledger, and analytics are feature packages in the same process. Do not add an API Gateway or split these features into independently deployed services unless the user explicitly requests a future architecture change.

Keep domain code organized by feature and keep HTTP handlers thin:

```text
backend/service/
├── cmd/api/
├── internal/
│   ├── auth/
│   ├── ledger/
│   │   ├── categories/
│   │   ├── wallets/
│   │   ├── transactions/
│   │   └── budgets/
│   ├── analytics/
│   ├── httpapi/
│   └── config/
├── migrations/
├── sql/
├── Dockerfile
└── go.mod
```

Analytics calls ledger application services in-process. Do not use HTTP to communicate between packages in this service.

## Backend Stack

Use:

- Go
- Echo
- pgx/v5
- sqlc
- goose
- go-playground/validator when needed
- JWT
- bcrypt or Argon2id if password authentication is introduced
- slog
- Docker

Do not add gRPC, Protocol Buffers, NATS, Redis, Kafka, or Kubernetes dependencies to the MVP unless the current task explicitly requires them.

## Database

Use two PostgreSQL databases configured with `AUTH_DATABASE_URL` and `LEDGER_DATABASE_URL`. Neon is the default persistent online database. Auth owns `auth_db`; the ledger feature owns `ledger_db`. The single Go process opens one connection pool to each database.

Use Goose migrations from `backend/service/migrations/auth/` and `backend/service/migrations/ledger/` for their respective database changes. Never manually edit a production schema. Keep SQL query definitions under `backend/service/sql/` and generated pgx code in the owning feature package.

Store monetary values using PostgreSQL `NUMERIC`, never floating point. Use UTC for backend timestamps. Use UUIDs for public entity identifiers.

## HTTP API

Expose `/health` for process health and `/ready` for database readiness. Public application routes use `/api/v1`:

```text
/api/v1/auth/*
/api/v1/transactions/*
/api/v1/categories/*
/api/v1/wallets/*
/api/v1/budgets/*
/api/v1/analytics/*
```

Keep the API contract stable and document public route changes. The frontend must use one public API URL and must not know internal package structure.

## Authentication and data isolation

JWT payloads stay minimal, for example `sub` and `exp`. Every request for user-owned data must validate the access token and scope database queries by the authenticated `user_id`. Never trust a user-provided ID for authorization.

Google OAuth is the current sign-in provider. Keep refresh tokens opaque, store only their hashes, and use secure, HTTP-only cookies with appropriate SameSite and path settings.

## Request flow

```text
HTTP Request
    -> Echo handler
    -> feature service
    -> repository
    -> sqlc / pgx
    -> PostgreSQL
```

Handlers parse and validate transport input, call feature services, and map results to HTTP responses. Services own business rules and orchestration. Repositories own persistence details. Do not create interfaces for every type; use them at real test or implementation boundaries.

## Local development and deployment

Use Docker Compose to run the single API container. The databases may be Neon or local PostgreSQL databases configured through `AUTH_DATABASE_URL` and `LEDGER_DATABASE_URL`. Deploy one backend service; do not deploy separate auth, ledger, analytics, or gateway processes.

Kubernetes is a later deployment target for the API if explicitly needed. Keep PostgreSQL on Neon unless self-hosting it becomes a deliberate objective.

## Coding principles

1. Keep code simple and explicit.
2. Avoid premature optimization and distributed-system complexity.
3. Keep feature packages cohesive and package dependencies directed inward.
4. Validate all external input.
5. Preserve user-level data isolation in every query.
6. Add indexes based on actual query patterns.
7. Use migrations for schema changes.
8. Avoid silent API contract changes.
9. Update relevant documentation when architecture or public APIs change.
10. Prefer small, reviewable changes.

If requirements are ambiguous, choose the simplest implementation consistent with the single-service architecture.
