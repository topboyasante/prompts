# Repository Structure

## Top Level

- `cmd/api`: API entrypoint and route wiring.
- `cmd/cli`: CLI entrypoint and command implementations.
- `internal`: application modules and infrastructure.
- `pkg/client`: shared HTTP client used by CLI.
- `migrations`: SQL schema migrations.

## Internal Modules

- `internal/auth`: OAuth login/callback, JWT issuance, auth middleware.
- `internal/users`: user model and data access.
- `internal/identities`: provider identity linking (`github`, `google`).
- `internal/prompts`: prompt create/get/search APIs.
- `internal/versions`: prompt version upload/list/download and tarball helpers.
- `internal/storage`: Cloudflare R2 client abstraction.
- `internal/db`: GORM DB open and migration runner.
- `internal/server`: common API middleware and response helpers.
- `internal/config`: environment variable loading.

## Boundary Rules

- `cmd/api` composes concrete implementations.
- Modules interact through repository/service interfaces.
- Cross-cutting response shape is centralized in `internal/server/respond.go`.
