# Prompts.dev

Prompts.dev is a registry and distribution platform for AI prompts.

This repository contains:
- a Go API server (`cmd/api`)
- a Go CLI (`cmd/cli`)
- database migrations (`migrations`)

## Quick Start

1. Copy env file:

```bash
cp .env.example .env
```

2. Update required values in `.env` (at minimum: `DATABASE_URL`, `JWT_SECRET`, OAuth credentials, R2 credentials).

Optional: boot local Postgres via Docker Compose:

```bash
docker compose up -d postgres
```

3. Run API server:

```bash
go run ./cmd/api
```

4. Use CLI:

```bash
go run ./cmd/cli --help
```

## Documentation

- Docs index: `docs/index.md`
- Contributor guide: `docs/contributing.md`
- Local development: `docs/local-development.md`
- API spec (OpenAPI): `docs/api/openapi.yaml`
- Frontend API guide: `docs/api/integration-guide.md`

## Current API Base URL

- Local: `http://localhost:8080/v1`

## Notes

- API error responses use a standard envelope and include `request_id`.
- OAuth supports `github` and `google` providers.
