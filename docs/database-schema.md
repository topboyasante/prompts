# Database Schema

## Tables

- `users`
  - local account record
  - columns: `id`, `username`, `email`, `avatar_url`, `created_at`

- `user_identities`
  - external provider mapping
  - columns: `id`, `user_id`, `provider`, `provider_user_id`, `email`, `email_verified`, `created_at`
  - unique: `(provider, provider_user_id)`
  - unique: `(user_id, provider)`

- `prompts`
  - prompt metadata per owner
  - unique: `(owner_id, name)`

- `prompt_versions`
  - version metadata and storage key
  - unique: `(prompt_id, version)`

- `prompt_tags`
  - tags for search/discovery

- `downloads`
  - append-only download events

## Extensions and Indexes

- `pgcrypto` for UUID generation.
- `pg_trgm` and trigram index for prompt name similarity search.
