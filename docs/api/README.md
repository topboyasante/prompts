# API Documentation

`openapi.yaml` is the source of truth for the HTTP API contract.

## Files

- `openapi.yaml`: machine-readable API spec.
- `integration-guide.md`: frontend usage notes and flow examples.
- `error-codes.md`: API error code reference.
- `llm-frontend-context.md`: LLM-friendly frontend build context.

## Validate Spec

Example using swagger-cli:

```bash
npx @apidevtools/swagger-cli validate docs/api/openapi.yaml
```
