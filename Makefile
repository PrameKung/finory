.PHONY: help backend-check frontend-path-test run-frontend run-api-gateway run-auth-service run-ledger-service run-analytics-service compose-up compose-down compose-logs

COMPOSE := docker compose --env-file .env.local -f deployments/docker/compose.yaml

help:
	@echo "Run 'make run-frontend' for Next.js."
	@echo "Run 'make run-api-gateway', 'make run-auth-service', 'make run-ledger-service', or 'make run-analytics-service' for Go services."
	@echo "Run 'make frontend-path-test' for the Docker-backed Gateway -> Ledger -> PostgreSQL integration test."
	@echo "Run 'make compose-up' to start the local backend stack, 'make compose-logs' to follow logs, or 'make compose-down' to stop it."

run-frontend:
	cd frontend && npm run dev

run-api-gateway run-auth-service run-ledger-service run-analytics-service:
	cd backend/services/$(patsubst run-%,%,$@) && go run ./cmd/api

backend-check:
	cd backend/services/api-gateway && go test ./...
	cd backend/services/auth-service && go test ./...
	cd backend/services/ledger-service && go test ./...
	cd backend/services/analytics-service && go test ./...

frontend-path-test:
	cd backend/integration && go test -tags=integration -run TestFrontendFacingTransactionPath -v

compose-up:
	$(COMPOSE) up --build -d --wait

compose-down:
	$(COMPOSE) down

compose-logs:
	$(COMPOSE) logs -f
