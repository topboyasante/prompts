# API Error Codes

All errors use the same envelope:

```json
{
  "error": {
    "code": "...",
    "message": "...",
    "request_id": "...",
    "details": {}
  }
}
```

## Codes

- `VALIDATION_FAILED`
  - Request payload, params, or query are invalid.
  - Check `error.details` for field-level hints.

- `UNAUTHORIZED`
  - Missing/invalid bearer token or OAuth state failures.

- `NOT_FOUND`
  - Requested prompt or version does not exist.

- `CONFLICT`
  - Resource already exists (prompt name conflict or duplicate version).

- `INTERNAL_ERROR`
  - Unexpected server-side failure.

## Frontend Handling

- Always surface `error.message` to logs.
- Capture and display/report `error.request_id` for support.
