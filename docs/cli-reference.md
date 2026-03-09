# CLI Reference

Binary name: `prompt`

## `prompt login`

Authenticate using OAuth.

```bash
prompt login --provider github
prompt login --provider google
```

Stores JWT at `~/.prompts/config.json`.

## `prompt init [name]`

Scaffold prompt files.

```bash
prompt init my-prompt
prompt init my-prompt --force
```

## `prompt publish`

Publish current package directory.

Flow:
- read and validate `prompt.yaml`
- create prompt metadata
- create tarball
- upload version

## `prompt install [owner/name[@version]]`

Install prompt to local `.prompts/{name}` directory.

```bash
prompt install topboy/landing-page-writer
prompt install topboy/landing-page-writer@1.0.0
```

## `prompt run [name] --var key=value`

Render local prompt template with variables.

```bash
prompt run landing-page-writer --var product="AI SaaS"
```

## `prompt search [query]`

Search published prompts.

```bash
prompt search landing
```
