# Catalogue Export

The Faultline CLI can generate a static-site-ready public failure catalogue from
the bundled playbooks. The catalogue is consumed by the Faultline Teams /
landing-page repository to render human-readable failure reference pages.

## Output structure

```
catalogue/
  failures/
    <slug>.md              — one Markdown file per failure
  catalogue.json           — full failure index (JSON array)
  catalogue.manifest.json  — provenance and generation metadata
```

### Failure Markdown files

Each `failures/<slug>.md` contains Astro content-collection frontmatter
followed by a structured Markdown body:

```markdown
---
title: "Human readable title"
description: "Short SEO-friendly description"
failure_id: "stable-failure-id"
category: "auth"
ecosystems: ["docker","github-actions"]
signals:
  - "pull access denied"
  - "unauthorized: authentication required"
confidence: "high"
---

# Human readable title

## What this failure means
...

## Symptoms
...
```

### catalogue.json

A JSON array of all entries, sorted by category then slug.  Each entry contains:

| Field          | Description                                     |
|----------------|-------------------------------------------------|
| `slug`         | URL-safe identifier derived from `failure_id`   |
| `title`        | Human-readable title                            |
| `description`  | First sentence of the playbook summary          |
| `failure_id`   | Stable playbook ID                              |
| `category`     | Failure category (auth, build, deploy, …)       |
| `ecosystems`   | Derived from playbook tags                      |
| `signals`      | Top 8 log-match patterns                        |
| `confidence`   | Derived from playbook severity                  |
| `source_path`  | Relative path to the source YAML playbook       |

### catalogue.manifest.json

```json
{
  "source_repo":       "org/faultline",
  "source_commit":     "abc123def456...",
  "generated_at":      "2025-01-01T00:00:00Z",
  "failure_count":     182,
  "generator_version": "1.2.3"
}
```

## Generating locally

### Make target (recommended)

```bash
# Generate catalogue/ in the repo root
make catalogue-export

# Generate and then validate output files exist
make catalogue-validate
```

### Direct CLI invocation

```bash
# Build the CLI first
make build

# Run the export
./bin/faultline catalogue export \
  --src playbooks/bundled \
  --out catalogue \
  --repo org/faultline \
  --commit "$(git rev-parse HEAD)"
```

All flags:

| Flag        | Default                              | Description                              |
|-------------|--------------------------------------|------------------------------------------|
| `--src`     | `playbooks/bundled`                  | Root of the playbook tree to scan        |
| `--out`     | `catalogue`                          | Output directory                         |
| `--repo`    | `$GITHUB_REPOSITORY` or `faultline`  | Source repo name stamped into manifest   |
| `--commit`  | `$GITHUB_SHA` or `git rev-parse HEAD`| Commit SHA stamped into manifest         |
| `--version` | *(empty)*                            | Generator version stamped into manifest  |

## GitHub Actions workflow

The `.github/workflows/catalogue-export.yml` workflow runs automatically when
playbook files or catalogue generator code change.  It:

1. Runs `go test ./internal/catalogue/...`
2. Builds the CLI with `make build`
3. Runs `faultline catalogue export` with the GitHub SHA and repo name
4. Validates that the expected output files are present
5. Uploads `catalogue.tar.gz` as a GitHub Actions artifact (30-day retention)

### Syncing into the Teams repository

The `sync-to-teams` job (in the same workflow) opens a pull request into the
`faultline/faultline-teams` repository when the `FAULTLINE_BOT_TOKEN` secret is
configured.  The PR copies:

- `catalogue/failures/*.md` → `docs-site/src/content/failures/`
- `catalogue/catalogue.json` → `docs-site/src/data/catalogue.json`
- `catalogue/catalogue.manifest.json` → `docs-site/src/data/catalogue.manifest.json`

The PR is titled **"Update generated failure catalogue"** and includes the source
commit SHA in the body.  All changes go through review before merging into the
Teams repo.

To enable the sync, add a `FAULTLINE_BOT_TOKEN` secret to the CLI repository
with a GitHub token that has write access to the Teams repository.

## Validation

The catalogue package enforces the following invariants:

- Slugs are URL-safe: lowercase letters, digits, and hyphens only
- Every entry must have a `failure_id`, `title`, and `description`
- Slugs within the same export must be unique
- `catalogue.json` and `catalogue.manifest.json` must be valid JSON

Validation runs automatically during export.  You can also call the validator
functions directly from the `internal/catalogue` package:

```go
if err := catalogue.ValidateEntries(entries); err != nil {
    // handle invalid entries
}
if err := catalogue.ValidateJSON(data); err != nil {
    // handle invalid JSON
}
```

## Package

Core logic lives in `internal/catalogue/catalogue.go`.  The CLI wiring is in
`internal/cli/cmd_catalogue.go`.  The `catalogue` command is a hidden
maintainer surface (`SurfaceMaintainer`) and does not appear in end-user help.
