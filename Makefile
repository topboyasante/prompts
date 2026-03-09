SHELL := /bin/bash

REQUIRED_ENV := DATABASE_URL JWT_SECRET R2_ACCOUNT_ID R2_BUCKET R2_ACCESS_KEY_ID R2_SECRET_ACCESS_KEY

.PHONY: help up down reset api cli build-cli install-cli test fmt tidy check openapi-validate env-check

help:
	@echo "Available targets:"
	@echo "  make up               - Start local Postgres"
	@echo "  make down             - Stop local services"
	@echo "  make reset            - Stop services and remove volumes"
	@echo "  make api              - Run API server"
	@echo "  make cli              - Run CLI help"
	@echo "  make build-cli        - Build CLI binary to ./bin/prompt"
	@echo "  make install-cli      - Install CLI binary into GOPATH/bin"
	@echo "  make test             - Run Go tests"
	@echo "  make fmt              - Format Go code"
	@echo "  make tidy             - Tidy Go modules"
	@echo "  make env-check        - Validate required runtime env vars"
	@echo "  make openapi-validate - Validate OpenAPI spec"
	@echo "  make check            - Run fmt, test, and OpenAPI validation"

up:
	docker compose up -d postgres

down:
	docker compose down

reset:
	docker compose down -v

api:
	go run ./cmd/api

cli:
	@set -a; [ -f .env ] && source .env || true; set +a; go run ./cmd/cli --help

build-cli:
	@mkdir -p bin
	go build -o ./bin/prompt ./cmd/cli

install-cli:
	go install ./cmd/cli

test:
	go test ./...

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./...)

tidy:
	go mod tidy

openapi-validate:
	npx @apidevtools/swagger-cli validate docs/api/openapi.yaml

env-check:
	@test -f .env || (echo ".env not found. Copy .env.example to .env and fill required values." && exit 1)
	@set -a; source .env; set +a; \
	for key in $(REQUIRED_ENV); do \
		if [ -z "$${!key}" ]; then \
			echo "Missing required env var: $$key"; \
			exit 1; \
		fi; \
	done

check: fmt test openapi-validate
