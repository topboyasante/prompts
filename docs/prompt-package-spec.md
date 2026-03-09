# Prompt Package Specification

A prompt package is a directory containing prompt metadata and content.

## Required Files

- `prompt.yaml`
- `prompt.md`

Optional:
- `README.md`

## Example Structure

```text
landing-page-writer/
  prompt.yaml
  prompt.md
  README.md
```

## `prompt.yaml`

Example:

```yaml
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

## `prompt.md`

Supports placeholders like `{{product}}` and `{{audience}}`.

The CLI `run` command replaces placeholders using `--var key=value` arguments.
