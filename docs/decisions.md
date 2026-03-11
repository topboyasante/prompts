# Architectural Decisions

> ADR entries explain WHY — not what was built, but why it was built that way.

---

## No ORM for Migrations
**Date:** 2026-03-10 (inferred from scan)
**Why:** SQL migration files (`migrations/*.sql`) are used instead of GORM AutoMigrate, giving full control over schema changes and making them reviewable as plain SQL.
**Tradeoffs:** More verbose than AutoMigrate; requires manual migration files.
**Alternatives considered:** GORM AutoMigrate (rejected — harder to track incremental changes in production).

---

## Repository Interface Per Domain
**Date:** 2026-03-10 (inferred from scan)
**Why:** Each domain package (`prompts`, `versions`, `users`, `identities`) defines its own `Repository` interface. Handlers depend on the interface, not the GORM struct.
**Tradeoffs:** More boilerplate; each domain needs an interface + concrete type.
**Alternatives considered:** Passing GORM `*gorm.DB` directly to handlers (rejected — breaks testability and layering).

---

## Dual-Binary Architecture (API + CLI in one repo)
**Date:** 2026-03-10 (inferred from scan)
**Why:** Keeps API contract and CLI client in sync. Shared Go types (e.g., `versions.CreateTarball`, `versions.ExtractTarball`) are reused by both binaries without a separate package.
**Tradeoffs:** CLI and API must be built and versioned from the same codebase.
**Alternatives considered:** Separate repos (rejected for MVP — extra coordination overhead).

---

## Download via Presigned URL Redirect (not proxy)
**Date:** 2026-03-10 (inferred from scan)
**Why:** API returns HTTP 302 to a 15-minute presigned R2 URL instead of proxying the tarball bytes. Keeps API servers stateless and offloads bandwidth to R2.
**Tradeoffs:** Clients must follow redirects; presigned URLs expire.
**Alternatives considered:** Proxy download through API (rejected — unnecessary bandwidth cost on API servers).

---

## Server-side Tarball Creation for Web Publish
**Date:** 2026-03-10
**Why:** Browser-side tar.gz creation (via fflate/js-zip) was slow and exposed technical complexity (YAML, semver) to non-technical users. Moving tarball creation to the API allows a single JSON call with just name, description, tags, and prompt text.
**Tradeoffs:** Version is always `1.0.0` for the web publish path; CLI can still publish arbitrary versions via the existing multipart endpoint.
**Alternatives considered:** Keep browser-side tarball (rejected — slow, exposes YAML format); add a separate "simple" endpoint that calls the existing two-step flow internally (rejected — same overhead, just hidden).

---

## CLI OAuth via Loopback Server
**Date:** 2026-03-10 (inferred from scan)
**Why:** CLI starts a local HTTP server on port 9876 to receive the OAuth callback token from the browser, matching the standard device-auth-like pattern for CLI tools.
**Tradeoffs:** Port 9876 must be free; requires the browser flow.
**Alternatives considered:** Device flow / polling (not implemented); manual token paste (worse UX).
