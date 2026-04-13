# Faultline

Faultline is a deterministic CLI for CI failure diagnosis.

It reads a raw CI log from a file or stdin, matches it against a library of
YAML playbooks, and returns a ranked list of likely failures with evidence and
concrete fix steps.

Faultline also supports modular detector playbooks. The built-in `inspect`
command runs source-aware playbooks against a repository tree using the same
deterministic output model used by log analysis.

Human-facing playbook guidance now lives in markdown block strings inside YAML:
structured fields decide the match, markdown explains the result.

The playbook layer is reviewed separately from runtime matching: bundled
playbooks are validated, then conflict-reported so overlapping patterns and
`match.none` exclusions stay deterministic as the catalog grows.

## Installation

For private distribution, download the latest release tarball from GitHub
Releases, then unpack and run Faultline from the extracted directory so the
bundled `playbooks/bundled/` directory stays adjacent to the binary.

```bash
curl -L <release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze build.log
```

If you move the binary elsewhere, also move `playbooks/bundled/` with it or
point `FAULTLINE_PLAYBOOK_DIR` at the bundled playbook directory.

## Quick start

```bash
# Analyze a log file
faultline analyze build.log

# Pipe from stdin
cat build.log | faultline analyze

# Detailed narrative output
faultline analyze --mode detailed build.log

# Detailed diagnosis with local git context
faultline analyze build.log --git --mode detailed

# Top 3 ranked candidates
faultline analyze --top 3 build.log

# Machine-readable JSON
faultline analyze --json build.log

# Machine-readable JSON enriched with recent repository context
faultline analyze --json --git --since 30d --repo . build.log

# Show only the fix steps
faultline fix build.log

# Inspect a repository for source-risk patterns
faultline inspect .
faultline inspect ./service --mode detailed
faultline inspect ./service --json

# Generate a local triage workflow from the top diagnosis
faultline workflow build.log

# Generate an agent-ready handoff prompt
faultline workflow build.log --mode agent --git --repo .

# Emit GitHub Actions annotations
faultline analyze --ci-annotations build.log

# Browse all playbooks
faultline list
faultline list --category auth

# Full detail for one playbook
faultline explain docker-auth
```

## Commands

| Command | Description |
|---------|-------------|
| `analyze [file]` | Ranked diagnosis from a file or stdin |
| `inspect [path]` | Ranked source-risk findings from a repository tree |
| `fix [file]` | Print only the fix steps for the top match |
| `workflow [file]` | Generate a deterministic local or agent-ready follow-up plan |
| `list` | List all built-in playbooks |
| `list --category <cat>` | Filter by category |
| `explain <id>` | Full detail for a single playbook |

## Validation And Review

Use the test suite to validate the CLI and the playbook loader:

```bash
make test
```

`make test` also exercises a noisy corpus of full-length CI logs in addition to
the synthetic bundled playbook fixtures, so release readiness is checked
against overlap-heavy real-world shapes rather than isolated pattern strings.

Track bundled playbook load and analysis cost as the catalog grows:

```bash
make bench
```

Run the full publish-grade validation path with one command:

```bash
make release-check VERSION=v0.1.0
```

This runs `make test`, `make review`, builds a release snapshot, and smoke tests
the packaged archive. Include Docker delivery validation when Docker is
available:

```bash
make release-check VERSION=v0.1.0 WITH_DOCKER=1 IMAGE=faultline-smoke
```

When the premium pack repository is available locally, include it explicitly so
the release path also fails on future cross-pack duplicate IDs or pack load
errors:

```bash
make premium-check PREMIUM_PACK_DIR=../faultline-premium-pack
make release-check VERSION=v0.1.0 PREMIUM_PACK_DIR=../faultline-premium-pack
```

For day-to-day local work, you can avoid repeating the path by creating the
ignored convenience symlink once:

```bash
make premium-link
make premium-check
make release-check VERSION=v0.1.0
```

`make premium-path` shows which local premium-pack directory Faultline will use.

Review the composed starter-plus-premium catalog with the same deterministic
pattern-conflict report used for bundled playbooks:

```bash
make premium-review PREMIUM_PACK_DIR=../faultline-premium-pack
```

This is a review harness, not a hard gate on shared patterns. Use it after
`make premium-check` to inspect overlap, exclusions, and ranking pressure across
the combined catalog.

Smoke test the packaged release archive before publishing it:

```bash
make release-snapshot VERSION=v0.1.0
make smoke-release VERSION=v0.1.0
```

Use the playbook review target to inspect exact overlapping patterns and
explicit exclusions before changing bundled rules:

```bash
make review
```

`premium-check` resolves premium packs in this order: `PREMIUM_PACK_DIR`, the
ignored local symlink at `playbooks/packs/premium-local`, a CI checkout at
`premium-pack`, then the sibling repository at `../faultline-premium-pack`.

Smoke test the Docker delivery path when Docker is available:

```bash
make docker-smoke IMAGE=faultline-smoke
```

Review guidance:

- Prefer tightening `match.any` or `match.all` before adding new exclusions.
- Add `match.none` only for high-confidence false positives that are shared
  with another rule.
- Keep source detector signals close to the risky scope so mitigations and
  suppressions can be interpreted structurally.
- Re-run `make review` after editing bundled playbooks to confirm the overlap
  report still makes sense.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--top N` | 1 | Show top N ranked results |
| `--mode quick\|detailed` | `quick` | Output verbosity |
| `--json` | false | Emit stable JSON instead of text |
| `--ci-annotations` | false | Emit `::warning::` GitHub Actions annotations |
| `--playbooks <dir>` | auto | Override playbook directory |
| `--playbook-pack <dir>` | none | Add an external pack root on top of bundled starter playbooks |
| `--no-history` | false | Skip reading and writing local history |
| `--git` | false | Enrich diagnoses with recent local git repository context |
| `--since <window>` | `30d` | Git history window for `--git`, for example `7d`, `2w`, `1 month ago` |
| `--repo <path>` | `.` | Repository path to scan when `--git` is enabled |

`inspect` supports `--top`, `--mode`, `--json`, `--playbooks`,
`--playbook-pack`, and `--no-history`.

## Output

Faultline auto-detects terminal output. ANSI styling and markdown rendering are
used only for interactive terminals; redirected output, CI, `NO_COLOR`, and
`FAULTLINE_PLAIN=1` fall back to readable plain text. `--json` always bypasses
terminal styling entirely.

**Quick mode** (default) — action-first, minimal noise:

```text
Docker registry authentication failure (docker-auth) [67% confidence]
Severity: high

Category: auth
Score: 1.00

Summary
-------
CI could not authenticate to the container registry before an image pull or push.
```

**Detailed mode** (`--mode detailed`) — compact diagnosis plus evidence and repo context:

```text
Docker registry authentication failure (docker-auth) [67% confidence]
Severity: high

Category: auth
Score: 2.00
Stage: build

Summary
-------
CI could not authenticate to the container registry before an image pull or push.

Evidence
--------
- pull access denied

Repo Context
------------
- Repo root: /workspace/app
- Recent file: .github/workflows/deploy.yml
- Related commit: 2026-04-10 a1b2c3d hotfix: restore registry login step
```

**JSON** (`--json`) — stable schema for automation:

```json
{
  "matched": true,
  "source": "build.log",
  "fingerprint": "a1b2c3d4",
  "context": { "stage": "build" },
  "repo_context": {
    "repo_root": "/workspace/app",
    "recent_files": [".github/workflows/deploy.yml", "Dockerfile"],
    "related_commits": [
      { "hash": "a1b2c3d", "subject": "hotfix: restore registry login step", "date": "2026-04-10" }
    ],
    "hotspot_directories": ["deploy"],
    "co_change_hints": ["Dockerfile <-> .github/workflows/deploy.yml"],
    "hotfix_signals": ["hotfix: restore registry login step"],
    "drift_signals": ["Repeated edits in deploy"]
  },
  "results": [
    {
      "rank": 1,
      "failure_id": "docker-auth",
      "title": "Docker registry authentication failure",
      "category": "auth",
      "severity": "high",
      "score": 2.00,
      "confidence": 0.67,
      "summary": "CI could not authenticate to the container registry before an image pull or push.",
      "diagnosis_markdown": "## Diagnosis\n\nDocker reached the registry, but the credential failed.",
      "fix_markdown": "## Fix steps\n\n1. Verify the registry secret.",
      "validation_markdown": "## Validation\n\n- Re-run the login and image pull.",
      "evidence": ["pull access denied"]
    }
  ]
}
```

## Playbook Authoring

Use yaml structure for detection and ranking, then markdown for human guidance.

```yaml
summary: |
  One-line operator summary.

diagnosis_markdown: |
  ## Diagnosis

  Short explanation of what failed and the most likely cause.

fix_markdown: |
  ## Fix steps

  1. Do the first thing.
  2. Do the second thing.

validation_markdown: |
  ## Validation

  - Re-run the failing command.
  - Confirm the error is gone.
```

See [docs/playbooks.md](docs/playbooks.md) for the contributor guide.

**Workflow** (`workflow`) — deterministic follow-up for local or agentic loops:

```text
WORKFLOW  docker-build-context · Docker build context or Dockerfile path issue  [agent · workflow.v1]
Source: build.log
Stage: build
Evidence:
  - failed to read Dockerfile
Focus files:
  - Dockerfile
  - .dockerignore
Local repro:
  - docker build -f Dockerfile .
Verify:
  - docker build -f Dockerfile .
Next steps:
  1. Confirm the top diagnosis `docker-build-context` by reproducing the failure from the same command or CI step if possible.
  2. Verify the exact `docker build` command and confirm the final argument points at the intended build context.

Agent prompt:
You are helping resolve a deterministic CI failure in the local repository.
Top diagnosis: docker-build-context - Docker build context or Dockerfile path issue.
...
```

## Playbooks

Playbooks are packaged as deterministic packs. The shipped starter catalog
lives under `playbooks/bundled/`:

```
playbooks/
  bundled/
    log/
      auth/      docker-auth, git-auth, missing-env
      build/     go-sum-missing, npm-ci-lockfile, yarn-lockfile
      test/      flaky-test, parallelism-conflict, order-dependency
      network/   network-timeout, ssl-cert-error, dns-resolution
      runtime/   oom-killed, permission-denied, disk-full
      deploy/    terraform-state-lock, health-check-failure, container-crash
Premium packs are distributed from a separate repository and loaded through
`--playbooks` or `FAULTLINE_PLAYBOOK_DIR`.

When extra packs are composed alongside the starter catalog, Faultline carries
pack provenance through `list`, `explain`, and analysis JSON so paid rules can
be audited and supported without guessing which catalog produced a match.
```

Each playbook is purely declarative:

```yaml
id: docker-auth
title: Docker registry authentication failure
category: auth
severity: high
base_score: 1.0
tags: [docker, registry, auth]
stage_hints: [build, deploy]

summary: |
  CI could not authenticate to the container registry before an image pull or push.

diagnosis_markdown: |
  ## Diagnosis

  Docker reached the registry, but the configured credential was missing,
  expired, or scoped incorrectly for the target image.

fix_markdown: |
  ## Fix steps

  1. Verify the registry username, token, or password configured in CI secrets.
  2. Ensure the registry login step runs before any docker pull or push.

validation_markdown: |
  ## Validation

  - Re-run the registry login step.
  - Confirm the image pull or push completes successfully.

why_it_matters_markdown: |
  ## Why it matters

  Registry auth failures block both build-time pulls and deploy-time pushes.

match:
  any:
    - pull access denied
    - authentication required
```

### Custom playbooks

Point `--playbooks` at any pack root (or set `FAULTLINE_PLAYBOOK_DIR`) and
Faultline will load all `.yaml` files found recursively:

```bash
faultline analyze --playbooks ./my-playbooks build.log
FAULTLINE_PLAYBOOK_DIR=/opt/playbooks/bundled faultline analyze build.log
```

To compose extra packs alongside the bundled starter catalog, repeat
`--playbook-pack` or set `FAULTLINE_PLAYBOOK_PACKS` using your platform path
separator:

```bash
faultline analyze --playbook-pack ./packs/acme --playbook-pack ./packs/team build.log
FAULTLINE_PLAYBOOK_PACKS=./packs/acme:./packs/team faultline analyze build.log
```

`faultline list` adds a `PACK` column for composed catalogs, and `faultline
explain <id>` includes the pack name when the playbook came from an external
pack.

`--playbooks` remains a full override and should not be combined with
`--playbook-pack`.

### Scoring

| Match type | Points |
|-----------|--------|
| `any` pattern hit | +1.0 per hit |
| `all` pattern hit | +1.5 per hit |
| All `all` patterns matched | +2.0 bonus |
| Stage hint matched | +0.75 |
| `base_score` (always) | playbook value |

Results are ranked by score descending, then confidence descending, then ID
alphabetically for stability.

## Workflow follow-up

Use `faultline workflow` when you want the diagnosis translated into a next
action plan rather than only a description.

- `faultline workflow build.log`
  Produces a local triage checklist with likely files, repro commands, and
  verification commands when the playbook defines them.
- `faultline workflow build.log --mode agent`
  Adds a structured agent prompt you can hand to a coding assistant while
  keeping Faultline itself deterministic.
- `faultline workflow --json`
  Emits the same plan in a stable automation-friendly JSON shape with a
  `schema_version` field for downstream tooling.

## Local history

Faultline writes the top result from each run to `~/.faultline/history.json`
(max 500 entries). On subsequent runs the `seen_count` field shows how many
times that failure was diagnosed before, helping teams track recurring issues.

Use `--no-history` to skip reading and writing history entirely.

## Repository context

When `--git` is enabled, Faultline uses the local git CLI only. It does not
call remote APIs or require a hosted service.

The repo context pass scans recent commits from the local repository and
surfaces practical hints that can help explain a failure:

- likely related recent files
- recent related commits
- churn hotspots by directory
- simple co-change hints
- hotfix-like and revert-like drift signals

Examples:

```bash
faultline analyze build.log --git
faultline analyze build.log --git --since 14d
faultline analyze build.log --git --since 1 month ago --repo ../my-service
faultline analyze build.log --json --git
```

## Docker

```bash
docker build -t faultline .

# Analyze a mounted log
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/build.log

# JSON output
docker run --rm -v "$(pwd)":/workspace faultline analyze --json /workspace/build.log
```

The image bundles the binary and playbook packs at `/playbooks`, and Faultline
loads `/playbooks/bundled` by default, so no extra configuration is needed.

## Development

```bash
make build    # build ./bin/faultline
make test     # go test ./...
make run LOG=build.log
```

## Architecture

```
cmd/                   Cobra command tree
internal/
  model/               Shared data types (Playbook, Result, Analysis, Context)
  playbooks/           Catalog, pack loading, YAML validation, review
  matcher/             Deterministic scoring and ranking
  engine/              Orchestration: load → normalize → match → history
  output/              All formatters: text, JSON, CI annotations
  cli/                 Input reader (file / stdin)
  app/                 Command handlers (RunAnalyze, RunFix, RunList, RunExplain)
  repo/                Local git scanning, history parsing, signals, correlation
playbooks/             Bundled starter catalog plus reserved pack boundaries
```

Design constraints:
- Deterministic substring matching only — same input always produces the same output.
- No external services, no network calls, no ML.
- Playbooks are pure data; no code execution from YAML.
- Engine is fast: a 50 000-line log typically completes in under 50 ms.

- file and stdin input paths behave the same
