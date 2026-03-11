# Prompts.dev

Prompts.dev is a registry and distribution platform for reusable AI prompts.

It gives teams a package-style workflow for prompts: define prompt metadata, publish versions, install them locally, and run them with variables. Think of it as prompt infrastructure inspired by npm-style package distribution.

## Why this project exists

Prompt workflows are usually fragmented across docs, chats, and code comments. That creates repeated work and no reliable version history.

Prompts.dev solves that by providing:
- a versioned prompt registry API
- an install-and-run developer CLI
- OAuth-authenticated publishing
- object-storage-backed prompt package distribution

## What is in this repo

- Go API (`cmd/api`) for auth, prompt metadata, version publishing, and downloads
- Go CLI (`cmd/cli`) for login, init, publish, install, run, and search
- Postgres migrations (`migrations`)
- API contract and contributor docs (`docs`)

## Core capabilities (current)

- OAuth login with `github` and `google`
- Prompt search, create, version listing, and version download redirect
- Prompt version upload as `tar.gz` to Cloudflare R2
- Standard API error envelope with `request_id`
- Request logging with structured JSON fields for debugging

## Local quick start

1. Create env file:

```bash
cp .env.example .env
```

2. Fill required values in `.env`:
- `DATABASE_URL`
- `JWT_SECRET`
- `R2_ACCOUNT_ID`, `R2_BUCKET`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`
- OAuth credentials (`GITHUB_*` and/or `GOOGLE_*`)

3. Start Postgres:

```bash
make up
```

4. Start API (runs migrations on startup):

```bash
make api
```

5. Build and use CLI:

```bash
make build-cli
./bin/prompt --help
```

## Example flow

```bash
# login
./bin/prompt login --provider github

# create a new prompt package
./bin/prompt init landing-page-writer

# publish from package directory
cd landing-page-writer
../bin/prompt publish

# install and run
cd ..
./bin/prompt install <owner>/landing-page-writer
./bin/prompt run landing-page-writer --var product="AI SaaS"
```

## Architecture at a glance

- API base URL (local): `http://localhost:8080/v1`
- Postgres stores users, identities, prompts, versions, tags, downloads
- Cloudflare R2 stores prompt tarball artifacts
- API returns 302 for version download to presigned object URL

## API contract for frontend and LLM tooling

- OpenAPI source of truth: `docs/api/openapi.yaml`
- Frontend integration guide: `docs/api/integration-guide.md`
- LLM frontend build context: `docs/api/llm-frontend-context.md`

## Documentation index

- Start here: `docs/index.md`
- Contributor guide: `docs/contributing.md`
- Local dev guide: `docs/local-development.md`

## Project status

This is an MVP codebase under active development. API and CLI flows are functional for local development, and contracts are documented for frontend implementation.
