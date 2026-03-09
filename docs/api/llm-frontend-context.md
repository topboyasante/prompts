# LLM Frontend Context

Use this file plus `docs/api/openapi.yaml` as the source context for generating the Prompts.dev web frontend.

## Product Goal

Build a web app where users can:
- authenticate with OAuth (`github` or `google`)
- search prompts
- view prompt details and versions
- publish prompts (authenticated)
- upload prompt versions (authenticated)

## API Base URL

- Local development: `http://localhost:8080/v1`

## Authentication Model

- OAuth routes:
  - `GET /auth/{provider}/login`
  - `GET /auth/{provider}/callback`
- Providers: `github`, `google`
- Protected endpoints require bearer JWT:

```http
Authorization: Bearer <jwt>
```

- Web callback returns JSON `{ "token": "..." }` and also sets `prompts_token` cookie.

## Response and Error Contract

- Success responses are route-specific JSON objects.
- Error responses are always:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "validation failed",
    "request_id": "uuid",
    "details": {
      "field": "reason"
    }
  }
}
```

- Every API response includes `X-Request-ID` header.
- Frontend should log both `X-Request-ID` and `error.request_id` when present.

## Core Endpoints for UI

- `GET /prompts?q=<query>&limit=<n>&offset=<n>`
  - Search list page.
- `GET /prompts/{owner}/{name}`
  - Prompt details page.
- `GET /prompts/{owner}/{name}/versions`
  - Versions list.
- `GET /prompts/{owner}/{name}/versions/{version}/download`
  - Returns `302` to presigned URL.

Authenticated:
- `POST /prompts`
- `POST /prompts/{id}/versions` (multipart)

## Data Shapes Used Most in UI

Prompt:

```json
{
  "id": "uuid",
  "name": "landing-page-writer",
  "description": "Generates landing page copy",
  "owner_id": "uuid",
  "owner_username": "topboyasante",
  "tags": ["marketing", "writing"],
  "created_at": "2026-03-08T20:00:00Z"
}
```

Prompt version:

```json
{
  "id": "uuid",
  "prompt_id": "uuid",
  "version": "1.0.0",
  "tarball_url": "owner-id/name/1.0.0.tar.gz",
  "created_at": "2026-03-08T20:10:00Z"
}
```

## Frontend Behavior Rules

- Always handle `4xx/5xx` with `error.code` branching.
- Preserve and display validation `error.details` near fields.
- For download endpoint, allow redirects (do not expect JSON body).
- For publish flow:
  1. create prompt via `POST /prompts`
  2. if `201`, upload version
  3. if `409` on prompt create, prompt likely already exists

## Suggested Frontend Pages

- Login page (provider buttons)
- Search page
- Prompt detail page (metadata + versions)
- Publish page (prompt metadata + tarball upload)

## Suggested LLM Prompt

Use this prompt with an LLM coding assistant:

```text
Build a production-quality frontend for Prompts.dev.

API contract source of truth: docs/api/openapi.yaml
Additional behavioral context: docs/api/llm-frontend-context.md

Requirements:
1) Implement auth (github/google), search, prompt details, versions list, and publish flow.
2) Follow the exact API routes and schemas from OpenAPI.
3) Implement robust error handling using error.code, error.details, and request IDs.
4) Handle download redirect endpoint correctly.
5) Keep components modular and typed.
```
