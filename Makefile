.PHONY: help backend-check frontend-path-test run-frontend run-api compose-up compose-down compose-logs

COMPOSE := docker compose --env-file .env.local -f deployments/docker/compose.yaml

help:
	@echo "Run 'make run-frontend' for Next.js or 'make run-api' for the Go API."
	@echo "Run 'make frontend-path-test' for the Docker-backed API -> PostgreSQL integration test."
	@echo "Run 'make compose-up' to start the local stack, 'make compose-logs' to follow logs, or 'make compose-down' to stop it."

run-frontend:
	cd frontend && npm run dev

run-api:
	cd backend/service && go run ./cmd/api

backend-check:
	cd backend/service && go test ./...

frontend-path-test:
	cd backend/integration && go test -tags=integration -run TestFrontendFacingTransactionPath -v

compose-up:
	$(COMPOSE) up --build -d --wait

compose-down:
	$(COMPOSE) down

compose-logs:
	$(COMPOSE) logs -f
