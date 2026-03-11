# Architecture

## Project Type
Go 1.25 + Next.js 16 — Monorepo (REST API + CLI under `apps/api/`; web frontend under `apps/web/`; reserved `packages/` and `infra/` for future use)

## Directory Map
```
prompts/
├── apps/
│   ├── api/                 ← Go application root
│   └── web/                 ← Next.js 16 frontend (App Router, TypeScript, Tailwind)
│       ├── lib/
│       │   ├── api.ts       ← Axios client + typed API functions
│       │   └── types.ts     ← API response type definitions
│       ├── components/
│       │   ├── prompt-card.tsx  ← Display card (name, description, tags)
│       │   └── search-bar.tsx   ← useQuery search with debounce + spinners
│       └── .env.local       ← NEXT_PUBLIC_API_URL
│       ├── cmd/
│       │   ├── api/main.go  ← API entry point
│       │   └── cli/main.go  ← CLI entry point
│       ├── internal/
│       │   ├── auth/        ← OAuth handlers, JWT, middleware
│       │   ├── config/      ← Env-based config loading
│       │   ├── db/          ← GORM connection + migrations runner
│       │   ├── identities/  ← OAuth provider identity model/repo
│       │   ├── prompts/     ← Prompt model, repo, HTTP handler
│       │   ├── server/      ← Gin setup, middleware, response helpers
│       │   ├── storage/     ← R2/S3 storage interface + implementation
│       │   ├── users/       ← User model/repo
│       │   └── versions/    ← Version model, repo, tarball, HTTP handler
│       ├── migrations/      ← SQL migrations + embed.go
│       ├── pkg/
│       │   └── client/      ← HTTP client used by CLI to call the API
│       ├── go.mod
│       └── go.sum
├── packages/                ← Reserved for future shared packages
├── infra/                   ← Reserved for future infrastructure code
├── docs/
├── Makefile                 ← Root Makefile delegating to apps/api/
└── README.md
```

## Module Overview
| Module/Package | Purpose |
|---|---|
| `cmd/api` | Wires dependencies, registers routes, starts Gin server |
| `cmd/cli` | Cobra CLI: login, init, publish, install, run, search |
| `internal/auth` | OAuth2 login/callback, HMAC state signing, JWT issue/verify |
| `internal/config` | Loads env vars into `Config` struct; validates required keys |
| `internal/db` | Opens GORM Postgres connection, runs embedded migrations |
| `internal/identities` | `user_identities` table — links OAuth provider IDs to users |
| `internal/prompts` | `prompts` + `prompt_tags` tables; CRUD + trigram search |
| `internal/server` | Gin engine factory, request-ID/logger/CORS/recovery middleware |
| `internal/storage` | `storage.Client` interface + Cloudflare R2 implementation (AWS SDK v2) |
| `internal/users` | `users` table — upsert by provider identity |
| `internal/versions` | `prompt_versions` + `downloads` tables; upload/list/download + tarball helpers |
| `migrations` | SQL files embedded via `embed.go`, run at startup via golang-migrate |
| `pkg/client` | Reusable HTTP client (reads token from `~/.prompts/config.json`) |

## Data Flow

**Publish flow (CLI → API → R2 → Postgres):**
1. CLI reads `prompt.yaml`, creates tarball via `versions.CreateTarball`
2. Calls `POST /v1/prompts` (creates metadata) then `POST /v1/prompts/:id/versions` (uploads tarball)
3. API stores tarball in R2 at key `{ownerID}/{name}/{version}.tar.gz`
4. API persists version record in Postgres with `tarball_url` = R2 key

**Download flow (CLI → API → R2 presigned URL):**
1. CLI calls `GET /v1/prompts/:owner/:name/versions/:version/download`
2. API generates 15-min presigned URL from R2 key and responds with HTTP 302
3. CLI follows redirect, writes tarball to `.prompts/{name}/`

**Auth flow (Browser + CLI loopback):**
1. CLI starts local HTTP server on port 9876, opens browser to `/v1/auth/:provider/login?cli=true`
2. API performs OAuth redirect → callback → upserts user → issues JWT
3. For CLI flow, API redirects to `localhost:9876/callback?token=...`
4. CLI stores token to `~/.prompts/config.json`

## External Dependencies
| Name | Purpose |
|---|---|
| `github.com/gin-gonic/gin` | HTTP router and middleware framework |
| `gorm.io/gorm` + `gorm.io/driver/postgres` | ORM and Postgres driver |
| `github.com/golang-migrate/migrate/v4` | SQL migration runner (embedded files) |
| `github.com/golang-jwt/jwt/v5` | JWT issue and verification |
| `golang.org/x/oauth2` | OAuth2 flows for GitHub and Google |
| `github.com/aws/aws-sdk-go-v2/service/s3` | Cloudflare R2 uploads and presigned URLs |
| `github.com/spf13/cobra` | CLI command framework |
| `github.com/spf13/viper` | Configuration (indirect; used by cobra ecosystem) |
| `github.com/sirupsen/logrus` | Structured JSON logging |
| `github.com/google/uuid` | UUID generation for request IDs |
| `github.com/joho/godotenv` | `.env` file loading |
| `gopkg.in/yaml.v3` | Parsing `prompt.yaml` manifest in CLI |
| `axios` (web) | HTTP client; `withCredentials: true`; error interceptor |
| `@tanstack/react-query` (web) | `useQuery` / `useMutation` with caching and `invalidateQueries` |
