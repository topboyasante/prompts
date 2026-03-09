# Local Development

## Prerequisites

- Go 1.25+
- PostgreSQL
- Cloudflare R2 credentials
- OAuth app credentials for GitHub and/or Google

## Setup

1. Copy env file:

```bash
cp .env.example .env
```

2. Fill required values in `.env`:
- `DATABASE_URL`
- `JWT_SECRET`
- `R2_*`
- `GITHUB_*` and/or `GOOGLE_*`

3. Install dependencies:

```bash
go mod tidy
```

## Local Infrastructure (Docker Compose)

Start Postgres:

```bash
docker compose up -d postgres
```

Stop Postgres:

```bash
docker compose down
```

Remove Postgres volume data:

```bash
docker compose down -v
```

Default compose DB credentials:
- user: `prompt`
- password: `prompt`
- database: `promptsdev`
- host: `localhost:5432`

## Run API

```bash
go run ./cmd/api
```

On startup, migrations are applied automatically.

## Run CLI

```bash
go run ./cmd/cli --help
```

Examples:

```bash
go run ./cmd/cli login --provider github
go run ./cmd/cli init my-prompt
```

## Troubleshooting

- **Migration errors:** verify `DATABASE_URL` and DB permissions.
- **OAuth provider unsupported:** ensure provider credentials are configured.
- **R2 client init fails:** verify `R2_ACCOUNT_ID`, bucket, key ID, and secret.
- **401 responses:** confirm JWT was stored by CLI in `~/.prompts/config.json`.
