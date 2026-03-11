# Patterns

## Naming Conventions
- Files: `snake_case.go` (e.g., `handler.go`, `repository.go`, `tarball.go`)
- Types/structs: `PascalCase` (e.g., `Prompt`, `AuthService`, `Handler`)
- Functions/methods: `PascalCase` for exported, `camelCase` for unexported
- Variables: `camelCase`
- Constants: `camelCase` for unexported (e.g., `requestIDKey`, `loggerKey`)

## Folder Conventions
- `cmd/{binary}/main.go` — entry point per binary; wires everything together
- `internal/{domain}/` — one package per domain entity, each containing `model.go`, `repository.go`, `handler.go` as applicable
- `pkg/client/` — code shared with CLI (not internal); the HTTP client for talking to the API
- `migrations/` — SQL files named `{NNN}_{desc}.up.sql` / `.down.sql`, embedded via `embed.go`

## Recurring Code Patterns

### Error handling
- Handlers always check errors and call a `server.RespondError(c, status, code, message)` helper — no raw `c.JSON` for errors
- Errors bubble up with structured logrus fields before responding
- The API error envelope pattern: `{ "error": { "code": "...", "message": "...", "request_id": "..." } }`

### API response envelope
- All success responses use `{ "data": <payload> }` via `server.RespondJSON`
- Single resources: `{ "data": <resource fields> }`
- Collections: `{ "data": { "items": [...], <meta> } }`
- Frontend `api.ts` functions unwrap `.data.data` to return the payload directly to callers

### Repositories
- Each domain has a `Repository` interface defined in the package
- Concrete implementation is `GORMRepository` (e.g., `users.NewGORMRepository(db)`)
- Handlers depend on the interface, not the concrete type — enables future mock testing

### Dependency injection
- Manual constructor injection throughout — no DI framework
- `main.go` is the composition root: creates repos → services → handlers → registers routes

### Auth context
- `auth.UserIDFromGin(c)` extracts `user_id` from Gin context (set by middleware)
- Handlers that need auth check `ok` and return 401 if false

### Validation
- Input validated at handler level before calling repo
- Slug validation: `^[a-z0-9-]+$` regex on prompt names
- Version validation: `^\d+\.\d+\.\d+$` semver regex
- Required fields checked with `strings.TrimSpace`

### Async / Concurrency
- CLI login uses a goroutine for the loopback HTTP server + `select` on `tokenCh` / `errCh` / timeout
- No concurrency patterns in the API itself — standard synchronous Gin handlers

## Testing Conventions
- No test files present in current codebase (not determinable from scan)
- Test file location would follow Go convention: `*_test.go` in same package

## Anti-Patterns Observed
- `cmd/cli/main.go` is a single large file (~430 lines) with all command logic inline — a candidate for future refactoring into sub-packages, but functional for MVP scope
