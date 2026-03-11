# Changelog

## 2026-03-11 — Delete prompt + browse-all + /me fixes
- Added `DELETE /v1/prompts/:id` — auth required; 403 if not owner; 204 on success
- Added `GET /v1/me` — registered under authorized group (was missing from routes)
- Added `migrations/002_downloads_cascade` — adds `ON DELETE CASCADE` to `downloads.prompt_id` FK so deleting a prompt doesn't violate FK constraint
- Added `Repository.Delete` and `Repository.List` methods to `internal/prompts`
- Fixed `Handler.Search`: empty `q` now calls `repo.List` (returns all prompts, newest first) instead of returning 400
- Fixed `Handler.Get`: was wrapping prompt in `gin.H{"prompt": prompt}` causing response shape mismatch; now passes prompt directly to `RespondJSON`
- Added `deletePrompt(id)` to `apps/web/lib/api.ts`
- Detail page (`[owner]/[name]/page.tsx`): owner-only Delete button with confirm dialog, spinner, and redirect to `/` on success
- `server.go` CORS: added `DELETE` to `Access-Control-Allow-Methods`
- Affected modules: `apps/api/internal/prompts`, `apps/api/internal/server`, `apps/api/cmd/api`, `apps/api/migrations`, `apps/web/lib`, `apps/web/app`

## 2026-03-10 — Simplified publish + download counts
- Added `POST /v1/prompts/publish` — single JSON endpoint that creates prompt, builds tar.gz server-side, uploads to R2, persists version (always `1.0.0`)
- Added `CreateTarballFromContent` to `versions/tarball.go` — in-memory tar.gz from name/description/content strings
- Added `DownloadCount` (read-only subquery field) to `Prompt` model; all Search and FindByOwnerAndName queries now return download_count
- Replaced complex publish form (author, version, inputs, README, tarball) with 4-field form: name, description, tags, prompt text
- Added `publishPrompt()` to `apps/web/lib/api.ts`
- Added `download_count` to `Prompt` type in `apps/web/lib/types.ts`
- `PromptCard` and detail page (`[owner]/[name]/page.tsx`) now display download count
- Affected modules: `apps/api/internal/versions`, `apps/api/internal/prompts`, `apps/api/cmd/api`, `apps/web/lib`, `apps/web/app`, `apps/web/components`

## 2026-03-10 — Standardize API responses + build detail and create pages
- Wrapped `RespondJSON` in `{ "data": v }` — all success responses now use `{ "data": <payload> }` envelope
- Fixed `prompts.Handler.Get` to pass prompt directly (was `gin.H{"prompt": prompt}`)
- Added `owner_username?` to `Prompt` type; fixed `ApiError` shape (`request_id` inside `error`)
- Fixed `api.ts` functions to unwrap `{ data: ... }` envelope; updated return types for all endpoints
- Fixed `SearchBar`: `data?.items` (not `data?.data`); added `enabled: query.length > 0`; "Type to search prompts" empty state
- Added `<Link>` wrapper to `PromptCard` when `owner_username` is set
- Created `app/[owner]/[name]/page.tsx` — prompt detail page with versions list and download links
- Created `app/[owner]/[name]/loading.tsx` — skeleton for detail route
- Created `app/new/page.tsx` — create prompt form with validation, mutation, 401 handling
- Affected modules: `apps/api/internal/server`, `apps/api/internal/prompts`, `apps/web/lib`, `apps/web/components`, `apps/web/app`

## 2026-03-10 — Axios + TanStack Query + home page with search
- Replaced fetch-based `lib/api.ts` with single axios instance (`withCredentials: true`, response error interceptor)
- Added `@tanstack/react-query` for all data fetching with `useQuery` + `keepPreviousData`
- Created `app/providers.tsx` — `QueryClientProvider` wrapper (`'use client'`)
- Updated `app/layout.tsx` — wraps children with `<Providers>`, fixed metadata title/description
- Created `components/prompt-card.tsx` — display card (name, description, tag badges)
- Created `components/search-bar.tsx` — debounced search (300ms), `isFetching` spinner, `isLoading` skeletons
- Rewrote `app/page.tsx` — home page with header, login buttons (GitHub/Google), `<SearchBar />`
- Created `app/loading.tsx` — route-level skeleton while page JS loads
- Affected modules: `apps/web/lib`, `apps/web/app`, `apps/web/components`

## 2026-03-10 — Browser auth + typed API client
- Fixed CORS: restricted `Access-Control-Allow-Origin` to `WEB_ORIGIN` (default `http://localhost:3000`) with `credentials: true`
- Auth middleware now falls back to `prompts_token` HTTP-only cookie when no `Authorization` header is present
- OAuth callback (web flow) now redirects to `WebOrigin + "/"` instead of returning JSON
- Added `WebOrigin` field to `Config`, loaded from `WEB_ORIGIN` env var
- `server.New()` now accepts `webOrigin string` parameter; `main.go` passes `cfg.WebOrigin`
- Added `apps/web/lib/types.ts` — TypeScript types for `User`, `Prompt`, `PromptVersion`, `ApiError`
- Added `apps/web/lib/api.ts` — typed fetch client using `credentials: 'include'`
- Added `apps/web/.env.local` with `NEXT_PUBLIC_API_URL`
- Affected modules: `internal/config`, `internal/server`, `internal/auth`, `cmd/api`, `apps/web/lib`

## 2026-03-10 — Monorepo restructure
- Moved Go application from repo root into `apps/api/` (cmd, internal, pkg, migrations, go.mod, go.sum, .env.example, docker-compose.yml)
- Created `packages/` and `infra/` placeholder directories for future use
- Replaced root `Makefile` with delegating version that runs all Go commands via `cd apps/api/`
- Module name `github.com/topboyasante/prompts` unchanged; no source file edits required
- Updated `docs/architecture.md` and `docs/implementation.md` to reflect new paths

## 2026-03-10 — Initial index
- First codebase scan
- Generated architecture.md, implementation.md, patterns.md, decisions.md, changelog.md
- Modules present: auth, config, db, identities, prompts, server, storage, users, versions
- Both API and CLI binaries functional
