# Implementation

## Entry Points

| File | Role |
|---|---|
| `apps/api/cmd/api/main.go` | Loads config, opens DB, runs migrations, wires repos/services/handlers, starts Gin on `PORT` |
| `apps/api/cmd/cli/main.go` | Loads `.env` from parent dirs, builds Cobra root command with 6 subcommands |

## Per-Module Breakdown

### `internal/config`
- **Entry point:** `apps/api/internal/config/config.go`
- **Key functions:** `config.Load()` — reads env vars, validates required keys (`DATABASE_URL`, `JWT_SECRET`), returns `*Config`. Includes `WebOrigin` (default `http://localhost:3000`) loaded from `WEB_ORIGIN`.
- **Non-obvious logic:** Walks up parent directories to find `.env` (CLI only, via `loadDotEnvFromParents`)

### `internal/db`
- **Entry point:** `apps/api/internal/db/db.go`
- **Key functions:** `db.Open(url)` → GORM handle; `db.RunMigrations(url, dir)` → runs embedded SQL files
- **Initialization:** Migrations embedded in `migrations/` via `//go:embed` and run at every startup (idempotent via migrate)

### `internal/auth`
- **Entry point:** `apps/api/internal/auth/handler.go`, `auth/service.go`, `auth/middleware.go`, `auth/token.go`
- **Key functions:**
  - `Handler.Login` — redirects to OAuth provider; sets signed HMAC state cookie
  - `Handler.Callback` — verifies state cookie, exchanges code, upserts user, issues JWT; for CLI flow redirects to `localhost:9876/callback`
  - `AuthService.UpsertUser` — finds or creates user by provider identity
  - `auth.NewMiddleware(secret).Authenticate()` — Gin middleware; extracts Bearer token or falls back to `prompts_token` cookie, sets `user_id` in context
- **Non-obvious logic:** State is HMAC-signed with `JWT_SECRET` to prevent CSRF. CLI login uses a loopback HTTP server (port `CLI_OAUTH_PORT`, default 9876) to capture the token from browser redirect.

### `internal/prompts`
- **Entry point:** `apps/api/internal/prompts/handler.go`
- **Key functions:**
  - `Handler.Create` — validates slug (`^[a-z0-9-]+$`), creates prompt + tags
  - `Handler.Get` — lookup by `owner/name`; responds with prompt directly (not nested under `"prompt"` key)
  - `Handler.Search` — if `q` is non-empty: trigram search via `pg_trgm`; if `q` is empty: calls `repo.List` (all prompts, ordered by `created_at DESC`)
  - `Handler.Delete` — auth required; 404 if not found; 403 if not owner; calls `repo.Delete`; returns 204
- **Key types:** `Prompt` (GORM model), `Repository` interface with `Create`, `FindByID`, `FindByOwnerAndName`, `List`, `Search`, `Delete`

### `internal/versions`
- **Entry point:** `apps/api/internal/versions/handler.go`, `versions/tarball.go`
- **Key functions:**
  - `Handler.Upload` — verifies prompt ownership, validates semver, uploads tarball to R2, persists version record
  - `Handler.Download` — records download, generates 15-min presigned R2 URL, returns HTTP 302
  - `versions.CreateTarball(dir)` — packs directory into `tar.gz` (used by CLI)
  - `versions.ExtractTarball(r, dest)` — unpacks tarball to destination (used by CLI install)
- **Non-obvious logic:** R2 key format is `{ownerID}/{name}/{version}.tar.gz`. Download is a redirect, not a proxy.

### `internal/storage`
- **Entry point:** `apps/api/internal/storage/interface.go`, `storage/r2.go`
- **Key types:** `storage.Client` interface with `Upload(ctx, key, body, size, contentType)` and `GetPresignedURL(ctx, key, expiry)`
- **Implementation:** `R2Client` uses AWS SDK v2 pointed at Cloudflare R2 endpoint (`https://{R2_ACCOUNT_ID}.r2.cloudflarestorage.com`)

### `internal/server`
- **Entry point:** `apps/api/internal/server/server.go`, `server/respond.go`
- **Key functions:**
  - `server.New(webOrigin string)` — returns Gin engine with requestID, recovery, logger (JSON via logrus), CORS middlewares
  - `server.RespondJSON`, `server.RespondError`, `server.RespondValidationError` — standard response envelope
- **Non-obvious logic:** Every request gets a UUID `request_id` injected into context and response header `X-Request-ID`. CORS is restricted to `webOrigin` with `credentials: true` (no wildcard).

### `apps/web/lib`
- **Entry point:** `apps/web/lib/api.ts`, `apps/web/lib/types.ts`
- **Key exports:**
  - `apiClient` — axios instance (`baseURL=NEXT_PUBLIC_API_URL`, `withCredentials: true`; error interceptor rejects with `err.response?.data`)
  - `searchPrompts`, `getPrompt`, `createPrompt`, `listVersions`, `uploadVersion`, `deletePrompt` — typed axios wrappers; return `r.data` directly
  - `deletePrompt(id)` — `DELETE /prompts/:id`; returns raw axios response (caller handles redirect/invalidation)
  - `getLoginURL(provider)` — returns redirect URL for OAuth login (browser navigation, not fetch)
  - `getDownloadURL` — returns URL string; browser follows the 302 to R2 directly
- **Non-obvious logic:** All requests use `withCredentials: true` so the `prompts_token` HTTP-only cookie is sent automatically. `uploadVersion` overrides `Content-Type` to `multipart/form-data` (axios sets multipart boundary automatically).

### `apps/web/app`
- **`providers.tsx`** — `'use client'` wrapper around `QueryClientProvider`; `QueryClient` created with `useState` so it's stable per mount
- **`layout.tsx`** — wraps `{children}` with `<Providers>`, sets metadata (`title: "Prompts"`)
- **`page.tsx`** — `'use client'` home page; renders header with login buttons + `<SearchBar />`
- **`loading.tsx`** — route-level skeleton (pulsing placeholders for header + search + cards)
- **`[owner]/[name]/page.tsx`** — `'use client'` prompt detail page; two parallel `useQuery` calls (prompt + versions); renders description, tags, versions list with download links; loading skeleton + not-found state; owner sees Delete button (confirm dialog → `useMutation(deletePrompt)` → invalidate `['prompts']` → redirect to `/`)
- **`[owner]/[name]/loading.tsx`** — static skeleton for detail route (header block, tags row, versions list)
- **`new/page.tsx`** — `'use client'` create prompt form; `useMutation(createPrompt)`; validates slug pattern; comma-split tags; 401 → specific auth error message; redirects to `/` on success

### `apps/web/components`
- **`prompt-card.tsx`** — pure display card; props: `{ prompt: Prompt }`; renders name, description, tag badges
- **`search-bar.tsx`** — `'use client'`; owns search state + display
  - Debounces input → `query` (300ms `useEffect` + `clearTimeout`)
  - `useQuery({ queryKey: ['prompts', query], placeholderData: keepPreviousData })` drives fetch
  - `isFetching` → spinning indicator inside input; `isLoading` (first load) → skeleton grid

### `pkg/client`
- **Entry point:** `apps/api/pkg/client/client.go`, `pkg/client/prompts.go`
- **Key functions:** `client.New()` — reads token from `~/.prompts/config.json`, builds HTTP client with Bearer auth
- **Methods:** `CreatePrompt`, `UploadVersion`, `GetVersions`, `DownloadVersion`, `SearchPrompts`

### `cmd/cli` commands
| Command | What it does |
|---|---|
| `login` | Starts loopback server, opens browser, saves JWT to `~/.prompts/config.json` |
| `init [name]` | Scaffolds `prompt.yaml`, `prompt.md`, `README.md` in a new directory |
| `publish` | Reads `prompt.yaml`, creates prompt via API, packs + uploads tarball |
| `install [owner/name[@version]]` | Downloads tarball, extracts to `.prompts/{name}/` |
| `run [name]` | Reads `.prompts/{name}/prompt.yaml` + `prompt.md`, substitutes `{{var}}` placeholders |
| `search [query]` | Calls API search, prints tabular results |

## Configuration
| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | API listen port |
| `APP_ENV` | `development` | Environment label |
| `DATABASE_URL` | — (required) | Postgres connection string |
| `JWT_SECRET` | — (required) | Signs JWTs and OAuth state cookies |
| `JWT_EXPIRY_HOURS` | `720` | JWT TTL in hours (30 days) |
| `GITHUB_CLIENT_ID/SECRET` | — | GitHub OAuth app credentials |
| `GITHUB_REDIRECT_URL` | `http://localhost:8080/v1/auth/github/callback` | GitHub callback URL |
| `GOOGLE_CLIENT_ID/SECRET` | — | Google OAuth app credentials |
| `GOOGLE_REDIRECT_URL` | `http://localhost:8080/v1/auth/google/callback` | Google callback URL |
| `R2_ACCOUNT_ID` | — | Cloudflare account ID |
| `R2_BUCKET` | — | R2 bucket name |
| `R2_ACCESS_KEY_ID` | — | R2 access key |
| `R2_SECRET_ACCESS_KEY` | — | R2 secret key |
| `CLI_OAUTH_PORT` | `9876` | Loopback port for CLI OAuth callback |
| `PROMPTS_API_URL` | `http://localhost:8080/v1` | API base URL used by CLI |
| `WEB_ORIGIN` | `http://localhost:3000` | Allowed CORS origin (must match frontend URL) |
