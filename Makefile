SHELL := /bin/bash
API_DIR := apps/api

REQUIRED_ENV := DATABASE_URL JWT_SECRET R2_ACCOUNT_ID R2_BUCKET R2_ACCESS_KEY_ID R2_SECRET_ACCESS_KEY

.PHONY: help up down reset api cli build-cli install-cli test fmt tidy check openapi-validate env-check

help:
	@echo "Available targets:"
	@echo "  make dev              - Run API + web together (Ctrl-C to stop both)"
	@echo "  make up               - Start local Postgres"
	@echo "  make down             - Stop local services"
	@echo "  make reset            - Stop services and remove volumes"
	@echo "  make api              - Run API server"
	@echo "  make web              - Run web dev server"
	@echo "  make cli              - Run CLI help"
	@echo "  make build-cli        - Build CLI binary to apps/api/bin/prompt"
	@echo "  make install-cli      - Install CLI binary into GOPATH/bin"
	@echo "  make test             - Run Go tests"
	@echo "  make fmt              - Format Go code"
	@echo "  make tidy             - Tidy Go modules"
	@echo "  make env-check        - Validate required runtime env vars"
	@echo "  make openapi-validate - Validate OpenAPI spec"
	@echo "  make check            - Run fmt, test, and OpenAPI validation"

up:
	cd $(API_DIR) && docker compose up -d postgres

down:
	cd $(API_DIR) && docker compose down

reset:
	cd $(API_DIR) && docker compose down -v

dev:
	@trap 'kill 0' SIGINT; \
	(cd $(API_DIR) && go run ./cmd/api) & \
	(cd apps/web && pnpm dev) & \
	wait

api:
	cd $(API_DIR) && go run ./cmd/api

web:
	cd apps/web && pnpm dev

cli:
	cd $(API_DIR) && set -a; [ -f .env ] && source .env || true; set +a; go run ./cmd/cli --help

build-cli:
	mkdir -p $(API_DIR)/bin
	cd $(API_DIR) && go build -o ./bin/prompt ./cmd/cli

install-cli:
	cd $(API_DIR) && go install ./cmd/cli

test:
	cd $(API_DIR) && go test ./...

fmt:
	cd $(API_DIR) && gofmt -w $$(go list -f '{{.Dir}}' ./...)

tidy:
	cd $(API_DIR) && go mod tidy

openapi-validate:
	npx @apidevtools/swagger-cli validate docs/api/openapi.yaml

env-check:
	@test -f $(API_DIR)/.env || (echo "$(API_DIR)/.env not found. Copy $(API_DIR)/.env.example to $(API_DIR)/.env and fill required values." && exit 1)
	@set -a; source $(API_DIR)/.env; set +a; \
	for key in $(REQUIRED_ENV); do \
		if [ -z "$${!key}" ]; then \
			echo "Missing required env var: $$key"; \
			exit 1; \
		fi; \
	done

check: fmt test openapi-validate
