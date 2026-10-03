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
- Use the project as a learning environment for scalable web architecture, PostgreSQL, microservices, Docker, and Kubernetes

The MVP should remain small and practical. Avoid adding infrastructure or abstractions that are not yet required.

---

## Repository Structure

Use a monorepo.

```text
expense-tracker/
├── frontend/
├── backend/
│   ├── services/
│   │   ├── api-gateway/
│   │   ├── auth-service/
│   │   ├── ledger-service/
│   │   └── analytics-service/
│   └── pkg/
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

The frontend must communicate only with the API Gateway. It must not know the internal addresses of individual microservices.

Example:

```text
Next.js
   |
   v
API Gateway
```

Use one public API base URL.

```env
NEXT_PUBLIC_API_URL=https://api.example.com
```

---

# Backend

## Architecture

The backend uses Go and a microservice architecture.

Services:

1. API Gateway
2. Auth Service
3. Ledger Service
4. Analytics Service

Do not split every entity into its own service.

In particular, do not create separate transaction, category, wallet, and budget microservices.

Those domains belong together inside the Ledger Service.

---

## Backend Stack

Use:

- Go
- Echo
- pgx/v5
- sqlc
- goose
- go-playground/validator
- JWT
- bcrypt or Argon2id
- slog
- Docker

Later, when justified:

- gRPC
- Protocol Buffers
- NATS JetStream
- OpenTelemetry
- Prometheus
- Kubernetes

Do not introduce gRPC, NATS, Redis, Kafka, or Kubernetes dependencies into the initial MVP unless the current task explicitly requires them.

---

# Microservices

## API Gateway

Responsibilities:

- Public backend entry point
- Route requests to internal services
- Authentication middleware where appropriate
- Request ID / tracing propagation
- Common CORS handling
- Common rate limiting later

The API Gateway has no database.

Example routing:

```text
/api/v1/auth/*          -> auth-service
/api/v1/transactions/*  -> ledger-service
/api/v1/categories/*    -> ledger-service
/api/v1/wallets/*       -> ledger-service
/api/v1/budgets/*       -> ledger-service
/api/v1/analytics/*     -> analytics-service
```

Only the API Gateway should normally be exposed publicly.

---

## Auth Service

Owns:

- User accounts
- Email/password authentication
- Password hashing
- Login
- Logout
- JWT / refresh-token logic
- Google OAuth later
- Password reset later
- Account deletion logic

Initial routes:

```text
POST /auth/register
POST /auth/login
POST /auth/logout
POST /auth/refresh
GET  /auth/me
```

The Auth Service owns `auth_db`.

No other service may directly query `auth_db`.

---

## Ledger Service

This is the core financial service.

Owns:

- Transactions
- Categories
- Wallets
- Budgets

Initial routes:

```text
GET    /transactions
POST   /transactions
GET    /transactions/:id
PATCH  /transactions/:id
DELETE /transactions/:id

GET    /categories
POST   /categories
PATCH  /categories/:id
DELETE /categories/:id

GET    /wallets
POST   /wallets
PATCH  /wallets/:id
DELETE /wallets/:id

GET    /budgets
POST   /budgets
PATCH  /budgets/:id
DELETE /budgets/:id
```

The Ledger Service owns `ledger_db`.

No other service may directly query `ledger_db`.

---

## Analytics Service

Responsibilities:

- Daily summary
- Monthly summary
- Yearly summary later
- Income / expense trends
- Category ranking
- Expense distribution
- Net balance
- Month-over-month comparison

Initial routes may include:

```text
GET /analytics/summary
GET /analytics/monthly
GET /analytics/categories
GET /analytics/trends
```

Initially, the Analytics Service does not require its own database.

It may request the required data from the Ledger Service and calculate the result.

Later, if analytics becomes event-driven or stores precomputed read models, create `analytics_db`.

Examples of data that would justify `analytics_db`:

- Precomputed monthly summaries
- Materialized analytics read models
- Event-derived reporting tables
- Aggregated dashboard data

---

# Database Strategy

Use Neon PostgreSQL for persistent online databases.

Use one Neon project initially.

Create separate PostgreSQL databases inside that Neon project.

Initial setup:

```text
Neon Project: expense-tracker
├── auth_db
└── ledger_db
```

Later, only when required:

```text
Neon Project: expense-tracker
├── auth_db
├── ledger_db
└── analytics_db
```

Do not use schemas as substitutes for service-owned databases unless there is a strong operational reason.

Each service must have its own database connection string.

Example:

```env
# auth-service
DATABASE_URL=postgresql://.../auth_db
```

```env
# ledger-service
DATABASE_URL=postgresql://.../ledger_db
```

A service must never directly access another service's database.

Bad:

```text
analytics-service -> ledger_db
```

Good:

```text
analytics-service -> Ledger API
```

Later:

```text
ledger-service -> NATS event -> analytics-service
```

---

# Local Development

Use Docker Compose.

Recommended local services:

```text
postgres
api-gateway
auth-service
ledger-service
analytics-service
```

NATS may be added later.

A single local PostgreSQL instance may contain multiple databases:

```text
postgres
├── auth_db
└── ledger_db
```

Production databases should remain on Neon initially.

Do not run production PostgreSQL in Kubernetes at the beginning.

Kubernetes should primarily run application services.

---

# Kubernetes Strategy

Kubernetes is a later deployment target for the Go services.

Recommended future architecture:

```text
Internet
   |
   v
Ingress
   |
   v
API Gateway
   |
   +--> Auth Service
   +--> Ledger Service
   +--> Analytics Service
            |
            v
          Neon
```

Do not move PostgreSQL into Kubernetes merely to make the system "more Kubernetes-native".

If self-hosted PostgreSQL on Kubernetes becomes a deliberate learning objective, use a PostgreSQL operator such as CloudNativePG rather than a simple PostgreSQL Deployment.

---

# Service Internal Structure

Each Go service should follow a similar structure.

Example:

```text
service/
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── config/
│   ├── server/
│   ├── middleware/
│   ├── database/
│   └── <feature>/
│       ├── handler.go
│       ├── service.go
│       ├── repository.go
│       ├── dto.go
│       └── model.go
│
├── migrations/
├── sql/
│   └── queries/
├── sqlc.yaml
├── Dockerfile
├── go.mod
└── go.sum
```

Use this request flow:

```text
HTTP Request
    |
    v
Handler
    |
    v
Service
    |
    v
Repository
    |
    v
sqlc / pgx
    |
    v
PostgreSQL
```

Responsibilities:

### Handler

- Parse HTTP input
- Validate transport-level input
- Call service methods
- Return HTTP status codes
- Return JSON responses

### Service

- Business rules
- Authorization rules related to the domain
- Transaction coordination
- Orchestration

### Repository

- Database access
- sqlc / pgx calls
- Persistence-specific behavior

### DTO

- Request and response structures

### Model

- Domain representation when needed

Do not create interfaces for every struct by default.

Create interfaces only when they provide real value, such as testing boundaries or multiple implementations.

---

# Shared Backend Packages

`backend/pkg/` is only for technical cross-service utilities.

Good examples:

```text
pkg/
├── logger/
├── tracing/
├── requestid/
└── response/
```

Do not place shared domain logic here.

Bad:

```text
pkg/
├── user/
├── transaction/
├── category/
└── wallet/
```

Microservices should not become tightly coupled through shared domain packages.

---

# Authentication

JWT payloads should stay minimal.

Example:

```json
{
  "sub": "user-uuid",
  "exp": 1790000000
}
```

Do not place transactions, categories, wallets, or other large user data inside JWTs.

Every request that accesses user-owned resources must enforce user-level data isolation.

A user must never be able to access another user's transactions, categories, wallets, or budgets.

Database queries for user-owned data should normally include the authenticated `user_id`.

Example:

```sql
SELECT *
FROM transactions
WHERE id = $1
  AND user_id = $2;
```

---

# Inter-Service Communication

Initial implementation:

```text
Frontend -> REST -> API Gateway
API Gateway -> REST -> Services
Services -> PostgreSQL
```

Use HTTP/REST first because it is easier to build, inspect, test, and debug.

Later, internal service communication may move to gRPC if there is a clear benefit.

Future event-driven communication may use NATS JetStream.

Example:

```text
Ledger Service
    |
    | TransactionCreated
    v
NATS
    |
    +--> Analytics Service
    +--> Notification Service
    +--> Audit Service
```

Do not add Kafka for the current scale.

---

# API Versioning

Use:

```text
/api/v1
```

Example:

```text
/api/v1/auth/login
/api/v1/transactions
/api/v1/categories
/api/v1/wallets
/api/v1/analytics/summary
```

---

# MVP Scope

Implement first:

1. User registration
2. User login
3. Authentication middleware
4. Transaction creation
5. Transaction list
6. Transaction update
7. Transaction deletion
8. Default categories
9. Custom categories
10. Wallet support
11. Monthly summary
12. Basic dashboard

Keep later features outside the initial implementation:

- OCR optimization
- Subscription
- Payment
- Advanced analytics
- Dark mode
- Complex notification systems
- Event-driven architecture
- Redis caching
- Kafka
- Full Kubernetes production deployment

---

# Implementation Order

Recommended order:

```text
1. Repository structure
2. Docker Compose for local development
3. Neon database setup
4. API Gateway skeleton
5. Auth Service
6. Ledger Service
7. Connect Next.js to API Gateway
8. Analytics Service
9. Integration tests
10. NATS / event-driven architecture later
11. Observability later
12. Kubernetes deployment later
```

Do not implement all services simultaneously.

Get one vertical flow working before expanding.

Example first vertical flow:

```text
Next.js Login
    ->
API Gateway
    ->
Auth Service
    ->
auth_db
```

Then:

```text
Next.js Create Transaction
    ->
API Gateway
    ->
Ledger Service
    ->
ledger_db
```

---

# Coding Principles

1. Keep code simple.
2. Prefer explicit code over hidden abstractions.
3. Avoid premature optimization.
4. Avoid premature distributed-system complexity.
5. Keep service boundaries strict.
6. Each service owns its own database.
7. Do not perform cross-service SQL joins.
8. Prefer SQL-first database access using sqlc.
9. Keep route handlers thin.
10. Keep business logic in services.
11. Keep persistence logic in repositories.
12. Validate all external input.
13. Never trust user-provided IDs without authorization checks.
14. Use UUIDs for public entity identifiers.
15. Store monetary values using PostgreSQL `NUMERIC`, never floating-point.
16. Use UTC for backend timestamps.
17. Format dates and currency for the user's locale only at the presentation layer.
18. Add indexes based on actual query patterns.
19. Write migrations for every schema change.
20. Never manually edit production database schemas.

---

# AI Agent Rules

When modifying this repository:

- Respect existing service boundaries.
- Do not move domain ownership between services without an explicit architectural decision.
- Do not add a new microservice just because a new entity exists.
- Do not allow one service to query another service's database.
- Do not add infrastructure dependencies without explaining why they are necessary.
- Prefer small, reviewable changes.
- Follow the existing directory conventions.
- Reuse existing components and utilities before creating new ones.
- Keep frontend pages thin and feature-oriented.
- Keep backend handlers thin.
- Use migrations for database changes.
- Add or update tests when changing important business behavior.
- Do not silently change API contracts.
- Update relevant documentation when architecture or public APIs change.
- Preserve backwards compatibility when practical.
- Avoid over-engineering.

If requirements are ambiguous, prefer the simplest implementation that preserves the architecture described in this file.

---

# Current Final Architecture

```text
                         Next.js
                            |
                            v
                       API Gateway
                            |
             +--------------+--------------+
             |              |              |
             v              v              v
        Auth Service   Ledger Service   Analytics
             |              |              |
             v              v              |
          auth_db        ledger_db          |
             \              /              /
              \            /              /
                  Neon PostgreSQL
```

Current databases:

```text
auth_db
ledger_db
```

Future optional database:

```text
analytics_db
```

The Analytics Service should not receive its own database until it actually needs persistent analytics-specific data.
