.PHONY: help backend-check

help:
	@echo "Run 'cd frontend && npm run dev' for Next.js."
	@echo "Run 'make run-api-gateway' (or auth-service, ledger-service, analytics-service) for Go services."

run-%:
	cd backend/services/$* && go run ./cmd/api

backend-check:
	cd backend/services/api-gateway && go test ./...
	cd backend/services/auth-service && go test ./...
	cd backend/services/ledger-service && go test ./...
	cd backend/services/analytics-service && go test ./...
