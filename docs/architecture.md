# Architecture

Prompts.dev uses a modular monolith architecture.

## Runtime Shape

- CLI communicates with API over HTTP.
- API persists metadata in Postgres.
- Tarball artifacts are stored in Cloudflare R2.

## API Lifecycle

1. Request enters Gin router (`/v1/...`).
2. Request ID middleware sets `X-Request-ID`.
3. Auth middleware validates bearer token on protected routes.
4. Handler executes module logic.
5. Structured JSON response is returned.
6. Request log is emitted with status and latency.

## Auth Architecture

- Provider-agnostic OAuth endpoints: `/v1/auth/{provider}/login|callback`.
- Supported providers: `github`, `google`.
- `users` table stores local user accounts.
- `user_identities` stores provider mappings.
- Auto-linking is enabled only when provider email is verified.

## Prompt Publishing Flow

1. Client creates prompt metadata via `POST /prompts`.
2. Client uploads tarball via `POST /prompts/{id}/versions`.
3. API stores tarball in R2 and stores key in `prompt_versions.tarball_url`.

## Prompt Install Flow

1. Client resolves version metadata.
2. Client calls download endpoint.
3. API records download and redirects (`302`) to presigned URL.
4. CLI downloads tarball and extracts into `.prompts/{name}`.
