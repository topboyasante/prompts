 ---

# Prompts.dev — Product Requirements Document (PRD)

**Version:** 1.0
**Date:** March 2026
**Status:** Implementation Ready
**Architecture:** Modular Monolith
**Primary Language:** Go
**CLI Language:** Go

---

# 1. Product Overview

## 1.1 Vision

Prompts.dev is a **registry and distribution platform for AI prompts**.

Developers should be able to:

* create prompts
* version prompts
* publish prompts
* install prompts via CLI
* run prompts locally
* share prompts publicly

The platform should function similarly to:

* GitHub for code
* npm for packages

But specifically for **AI prompts**.

---

# 2. Problem Statement

Prompts today are scattered across:

* documentation
* codebases
* note apps
* chat history

Problems:

* no version control
* no sharing mechanism
* no discoverability
* no standard prompt format
* no installation workflow

Developers repeatedly rewrite the same prompts.

---

# 3. Goals

## Primary Goals

Build a platform where developers can:

1. Publish prompts
2. Version prompts
3. Install prompts via CLI
4. Run prompts locally
5. Discover prompts

---

## Non-Goals (MVP)

The following are **not required for MVP**:

* prompt monetization
* prompt analytics
* prompt testing framework
* model benchmarking
* team organizations
* prompt editing in browser
* AI execution service

The CLI will simply **retrieve prompts** and **execute them locally**.

---

# 4. Core Concepts

## 4.1 Prompt

A prompt is a reusable instruction for an AI model.

Example:

```
Write a landing page for {{product}} targeting {{audience}}.
```

Prompts support **variables**.

---

## 4.2 Prompt Package

Prompts are distributed as **packages**.

Example structure:

```
landing-page-writer/

prompt.yaml
prompt.md
README.md
```

---

## 4.3 Prompt Registry

Central server storing:

* prompts
* versions
* metadata
* downloads

---

## 4.4 CLI

CLI allows developers to:

* initialize prompts
* publish prompts
* install prompts
* run prompts

---

# 5. User Personas

### Prompt Author

A developer who creates prompts and publishes them.

Needs:

* easy publishing
* versioning
* visibility

---

### Prompt Consumer

A developer using prompts in their project.

Needs:

* install prompts
* run prompts
* update prompts

---

# 6. System Architecture

```
CLI (Go)
   |
HTTP API
   |
Go Backend (Modular Monolith)
   |
Postgres Database
   |
Object Storage (S3/R2)
```

---

# 7. Backend Architecture

### Language

Go

---

### Framework

Recommended:

* `gin`
  or
* `chi`

---

### Storage

Database:

Postgres

Object Storage:

S3 compatible storage.

Stores prompt packages as:

```
tar.gz
```

---

# 8. Authentication

Authentication is done **only via GitHub OAuth**.

Using:

GitHub OAuth.

---

## Auth Flow

CLI:

```
prompt login
```

Steps:

1. CLI opens browser
2. User logs in via GitHub
3. Backend creates session
4. Backend returns JWT
5. CLI stores token

Token stored in:

```
~/.prompts/config.json
```

Example:

```
{
 "token": "jwt_token"
}
```

---

# 9. CLI Requirements

CLI binary name:

```
prompt
```

---

## 9.1 CLI Commands

### login

```
prompt login
```

Logs user in via GitHub.

---

### init

Creates prompt template.

```
prompt init my-prompt
```

Creates:

```
my-prompt/

prompt.yaml
prompt.md
README.md
```

---

### publish

Publishes prompt to registry.

```
prompt publish
```

Steps:

1. read `prompt.yaml`
2. validate structure
3. create tarball
4. upload to server
5. register version

---

### install

Installs prompt locally.

```
prompt install username/prompt-name
```

Downloaded to:

```
.prompts/
```

Example:

```
.prompts/landing-page-writer
```

---

### search

Search for prompts.

```
prompt search landing
```

---

### run

Runs prompt locally.

```
prompt run landing-page-writer --product="AI SaaS"
```

Steps:

1. load `prompt.md`
2. replace variables
3. output prompt text

Output printed to terminal.

---

# 10. Prompt Package Specification

Prompt packages follow a standard structure.

---

## Required Files

```
prompt.yaml
prompt.md
```

Optional:

```
README.md
```

---

## prompt.yaml Specification

Example:

```
name: landing-page-writer
description: Generates landing page copy
version: 1.0.0
author: topboy

inputs:
  - name: product
    required: true
  - name: audience
    required: false

tags:
  - marketing
  - writing
```

---

## prompt.md Example

```
You are a senior marketing copywriter.

Write a landing page for {{product}}.

Target audience: {{audience}}

Include sections:
- Hero
- Features
- Benefits
- CTA
```

---

# 11. Backend Modules

The backend is a **modular monolith**.

Modules:

```
auth
users
prompts
versions
storage
search
```

---

# 12. Database Schema

Database: Postgres

---

## users

```
users
```

Columns:

```
id UUID PRIMARY KEY
github_id TEXT UNIQUE
username TEXT
email TEXT
avatar_url TEXT
created_at TIMESTAMP
```

---

## prompts

```
prompts
```

Columns:

```
id UUID PRIMARY KEY
name TEXT
description TEXT
owner_id UUID
created_at TIMESTAMP
```

Unique constraint:

```
(owner_id, name)
```

---

## prompt_versions

```
prompt_versions
```

Columns:

```
id UUID PRIMARY KEY
prompt_id UUID
version TEXT
tarball_url TEXT
created_at TIMESTAMP
```

Unique constraint:

```
(prompt_id, version)
```

---

## prompt_tags

```
prompt_tags
```

Columns:

```
id UUID PRIMARY KEY
prompt_id UUID
tag TEXT
```

---

## downloads

```
downloads
```

Columns:

```
id UUID PRIMARY KEY
prompt_id UUID
version_id UUID
downloaded_at TIMESTAMP
```

---

# 13. API Endpoints

All endpoints return JSON.

---

## Auth

```
GET /auth/github/login
GET /auth/github/callback
```

Returns JWT.

---

## Create Prompt

```
POST /prompts
```

Body:

```
{
 "name": "landing-page-writer",
 "description": "Generates landing pages"
}
```

---

## Publish Version

```
POST /prompts/{id}/versions
```

Upload tarball.

---

## Search Prompts

```
GET /prompts?q=landing
```

---

## Get Prompt

```
GET /prompts/{owner}/{name}
```

---

## Get Versions

```
GET /prompts/{owner}/{name}/versions
```

---

## Download Version

```
GET /prompts/{owner}/{name}/versions/{version}/download
```

Returns tarball.

---

# 14. Repository Structure

Single repository containing CLI and API.

```
prompts/

cmd/
  api/
    main.go
  cli/
    main.go

internal/
  auth/
  users/
  prompts/
  versions/
  storage/

pkg/
  client/

migrations/

go.mod
```

---

# 15. CLI Implementation

CLI built using:

```
cobra
```

Commands:

```
prompt login
prompt init
prompt publish
prompt install
prompt run
prompt search
```

---

# 16. Storage

Prompt packages stored as:

```
tar.gz
```

Example:

```
landing-page-writer-1.0.0.tar.gz
```

Stored in object storage.

Example path:

```
prompts/{owner}/{name}/{version}.tar.gz
```

---

# 17. Prompt Installation

When installing:

```
prompt install topboy/landing-page-writer
```

CLI:

1. fetch latest version
2. download tarball
3. extract

Directory created:

```
.prompts/landing-page-writer
```

---

# 18. Prompt Execution

Running:

```
prompt run landing-page-writer --product="AI SaaS"
```

Steps:

1. read prompt.md
2. parse variables
3. replace with CLI args
4. print final prompt

Example output:

```
You are a senior marketing copywriter.

Write a landing page for AI SaaS.
```

---

# 19. Non Functional Requirements

### Performance

* prompt downloads < 500ms
* search < 200ms

---

### Security

* JWT authentication
* validate uploaded tarballs
* limit upload size

---

### Observability

Basic logging required.

Use:

```
logrus
```

---

# 20. Future Features (Not MVP)

Potential roadmap:

### Prompt testing

```
prompt test
```

---

### Prompt benchmarking

Compare outputs across models.

---

### Team workspaces

Organizations.

---

### Prompt monetization

Paid prompts.

---

### Prompt analytics

Track:

* downloads
* usage
* models used

---

# 21. Success Criteria (MVP)

MVP is successful if:

Users can:

1. login via GitHub
2. create prompts
3. publish prompts
4. install prompts via CLI
5. run prompts locally
6. discover prompts via search

---

# 22. Deliverables

The system must include:

* Go API server
* Go CLI
* Postgres migrations
* Prompt package validation
* GitHub OAuth login
* Prompt registry API
* CLI install and run functionality

---
