# Observability

## Request IDs

- Every response includes `X-Request-ID`.
- Error envelopes also include `error.request_id`.
- Use this ID to correlate logs and client errors.

## Logging

- API uses structured JSON logs.
- Logged fields include:
  - `request_id`
  - `method`
  - `path`
  - `status`
  - `latency_ms`
  - `user_id` (when authenticated)

## Error Contract

Errors follow this shape:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "validation failed",
    "request_id": "...",
    "details": {
      "field": "reason"
    }
  }
}
```
