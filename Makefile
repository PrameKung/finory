.PHONY: help backend-check run-frontend run-api-gateway run-auth-service run-ledger-service run-analytics-service

help:
	@echo "Run 'make run-frontend' for Next.js."
	@echo "Run 'make run-api-gateway', 'make run-auth-service', 'make run-ledger-service', or 'make run-analytics-service' for Go services."

run-frontend:
	cd frontend && npm run dev

run-api-gateway run-auth-service run-ledger-service run-analytics-service:
	cd backend/services/$(patsubst run-%,%,$@) && go run ./cmd/api

backend-check:
	cd backend/services/api-gateway && go test ./...
	cd backend/services/auth-service && go test ./...
	cd backend/services/ledger-service && go test ./...
	cd backend/services/analytics-service && go test ./...
