# Contributing Guide

## Workflow

- Create a branch from `main`.
- Keep PRs focused and small.
- Include tests or verification notes.
- Update docs when behavior or API contracts change.

## Required Checks

Run before opening a PR:

```bash
gofmt -w ./...
go test ./...
```

## API Change Policy

- Any API change must update `docs/api/openapi.yaml` in the same PR.
- Error envelopes must remain consistent with `internal/server/respond.go`.
- Keep routes under `/v1` unless introducing a breaking version.

## Commit Style

- Use concise, imperative commit messages.
- Explain why the change exists, not only what changed.

## Doc Style

- Prefer short sections and concrete examples.
- Use real request/response examples for API docs.
- Keep docs aligned with current implementation.
