# Frontend Integration Guide

## Base URL

- Local: `http://localhost:8080/v1`

## Auth Flow (Web)

1. Redirect user to `GET /auth/{provider}/login`.
2. Provider redirects back to `GET /auth/{provider}/callback`.
3. Callback sets `prompts_token` cookie and returns JSON token.

Providers:
- `github`
- `google`

## Auth Flow (CLI)

Pass `?cli=true` to login endpoint. Callback redirects with token query params to local CLI listener.

## Protected Endpoints

Send header:

```http
Authorization: Bearer <jwt>
```

Protected routes:
- `POST /prompts`
- `POST /prompts/{id}/versions`

## Download Endpoint Behavior

- `GET /prompts/{owner}/{name}/versions/{version}/download` returns `302` redirect to a presigned object URL.
- Frontend clients should allow redirect handling.

## Pagination

`GET /prompts` supports:
- `q` (required)
- `limit` (optional, default 20)
- `offset` (optional, default 0)

## Error Handling

- Parse `error.code` for branch logic.
- Record `error.request_id` in logs and support tooling.
