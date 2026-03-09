# Security Notes

## Authentication

- OAuth login supports `github` and `google`.
- JWT access tokens are issued after successful OAuth callback.
- Protected routes require `Authorization: Bearer <token>`.

## OAuth State Protection

- OAuth state is signed with HMAC.
- State payload includes provider name and nonce.
- Callback rejects invalid/mismatched state.

## Upload and Extraction Safety

- Version upload request body is capped at 10MB.
- Tar extraction rejects unsafe paths (`..` traversal, absolute paths).

## Sensitive Data

- Never log JWTs or OAuth tokens.
- Keep secrets in environment variables.
- Use strict file permission (`0600`) for `~/.prompts/config.json`.
