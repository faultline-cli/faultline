# Faultline

Faultline is a deterministic CLI for CI failure diagnosis.

It reads a CI log from a file or stdin, matches it against a reviewed playbook catalog, and returns the most likely failure with evidence pulled directly from the log plus concrete fix steps. The same runtime also supports repository inspection through explicit source-detector playbooks.

## Why This Repository Exists

Faultline is intentionally narrow:

- deterministic pattern matching only
- evidence-first output grounded in exact log lines
- local-first usage that also works cleanly in CI and Docker
- stable JSON and workflow output for automation and agents

The authoritative product and architectural constraints live in [SYSTEM.md](SYSTEM.md). This README is the operator and contributor entry point.

## Install

### Release archive

```bash
curl -L <release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze build.log
```

Each release archive contains:

```text
faultline_<version>_<os>_<arch>/
  faultline
  playbooks/
    bundled/
  README.md
```

If you move the binary elsewhere, keep `playbooks/bundled/` alongside it or set `FAULTLINE_PLAYBOOK_DIR` to the bundled catalog directory.

### Build from source

```bash
make build
./bin/faultline analyze build.log
```

## Quick Start

```bash
# Analyze a log file
faultline analyze build.log

# Read from stdin
cat build.log | faultline analyze

# Detailed human-readable output
faultline analyze build.log --mode detailed

# Markdown output for notes, issues, or agent handoff
faultline analyze build.log --format markdown

# Stable JSON for automation
faultline analyze build.log --json

# Add recent local git context
faultline analyze build.log --git --since 30d --repo .

# Show only the fix steps for the top diagnosis
faultline fix build.log

# Inspect a repository for source-risk findings
faultline inspect .

# Produce a deterministic follow-up workflow
faultline workflow build.log
faultline workflow build.log --mode agent --git --repo .

# Browse the active catalog
faultline list
faultline list --category auth
faultline explain docker-auth
```

## Command Surface

| Command | Description |
|---------|-------------|
| `analyze [file]` | Ranked diagnosis from a file or stdin |
| `inspect [path]` | Ranked source-risk findings from a repository tree |
| `fix [file]` | Print only the fix steps for the top diagnosis |
| `workflow [file]` | Build a deterministic local or agent-ready follow-up plan |
| `list` | List available playbooks from the active catalog |
| `explain <id>` | Show full detail for one playbook |
| `packs install <dir>` | Install an extra playbook pack into `~/.faultline/packs/` |
| `packs list` | Show installed extra playbook packs |

Common flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--top N` | 1 | Show top N ranked results |
| `--mode quick\|detailed` | `quick` | Human output verbosity for `analyze` and `inspect` |
| `--format raw\|markdown` | `raw` | Human output shape for `analyze`, `inspect`, `fix`, and `explain` |
| `--json` | false | Emit stable JSON |
| `--ci-annotations` | false | Emit GitHub Actions `::warning::` lines |
| `--playbooks <dir>` | auto | Replace the active catalog with one directory |
| `--playbook-pack <dir>` | none | Add one or more extra pack roots on top of the starter catalog |
| `--no-history` | false | Skip reading and writing local history |
| `--git` | false | Enrich diagnosis or workflow output with recent local git context |
| `--since <window>` | `30d` | History window for `--git` |
| `--repo <path>` | `.` | Repository path used by `--git` |

Environment variables:

| Variable | Description |
|----------|-------------|
| `FAULTLINE_PLAYBOOK_DIR` | Full catalog override, equivalent to `--playbooks` |
| `FAULTLINE_PLAYBOOK_PACKS` | Additional pack roots, separated by the platform path separator |
| `FAULTLINE_PLAIN=1` | Force plain text output even on interactive terminals |
| `NO_COLOR` | Disable ANSI colors |

## How It Works

The runtime stays deliberately explicit:

1. Read log input from a file path or stdin, or walk a repository tree for source inspection.
2. Normalize the input into stable lines or source signals.
3. Load the bundled starter catalog plus any installed or explicitly composed packs.
4. Validate playbook structure and review overlap conflicts deterministically.
5. Match with explicit detector and scorer logic.
6. Rank results with stable rules.
7. Optionally enrich with recent local git context.
8. Render concise text, markdown, JSON, or workflow output.

The main module boundaries are:

- `internal/cli`: Cobra command tree and flag handling
- `internal/app`: command-level use cases
- `internal/engine`: orchestration
- `internal/detectors`: explicit `log` and `source` detectors
- `internal/playbooks`: catalog loading, validation, and conflict review
- `internal/matcher`: evidence extraction and deterministic scoring
- `internal/output` and `internal/renderer`: JSON and human-facing rendering
- `internal/repo`: local git scanning and correlation
- `internal/workflow`: deterministic next-step plans

See [docs/architecture.md](docs/architecture.md) for the fuller architecture note.

## Catalog Model

Faultline ships with a bundled starter catalog under `playbooks/bundled/`.

The starter catalog favors broad first-run coverage: common auth, build, CI, deploy, network, runtime, and test failures across popular stacks, plus a minimal source-detector baseline so `faultline inspect .` is useful out of the box.

Premium or team-specific playbooks live outside the public starter release. Faultline composes them on top of the bundled catalog instead of baking them into the binary.

Recommended premium flow:

```bash
git clone <private-premium-pack-repo> ../faultline-premium-pack
faultline packs install ../faultline-premium-pack
faultline packs list
faultline list
```

Update flow:

```bash
cd ../faultline-premium-pack && git pull
faultline packs install --force ../faultline-premium-pack
```

One-off composition without installation:

```bash
faultline analyze build.log --playbook-pack ../faultline-premium-pack
```

Use `--playbooks <dir>` only for full catalog overrides. It replaces the starter catalog instead of composing with it.

## Output Modes

Faultline auto-detects terminal output. Styled rendering is used only for interactive terminals. Redirected output, CI, `NO_COLOR`, and `FAULTLINE_PLAIN=1` fall back to readable plain text. `--json` always bypasses terminal styling.

`--format raw` keeps the terminal-oriented output shape. `--format markdown` emits the same ranked diagnosis content as markdown source, which is useful for issue templates, handoff notes, and agent workflows.

Quick mode stays short and action-first. Detailed mode includes fuller evidence and repository context when available.

## Docker And CI

Build and run the starter image:

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/build.log
docker run --rm -v "$(pwd)":/workspace faultline analyze --json /workspace/build.log
```

The image bundles the starter catalog at `/playbooks/bundled` and prepares `/home/faultline/.faultline/packs` for installed extra packs.

To reuse the same installed packs locally and in Docker:

```bash
docker run --rm \
  -v "$(pwd)":/workspace \
  -v "$HOME/.faultline":/home/faultline/.faultline \
  faultline analyze /workspace/build.log
```

Common CI-oriented commands:

```bash
faultline analyze --json build.log
faultline analyze --ci-annotations build.log
faultline workflow build.log --mode agent --json
```

## Repository Map

```text
cmd/                  CLI entrypoints and maintenance commands
internal/app/         Command-level application services
internal/cli/         Cobra command tree
internal/engine/      Analysis orchestration
internal/detectors/   Log and source detector implementations
internal/playbooks/   Catalog loading, validation, composition, review
internal/matcher/     Evidence extraction and scoring
internal/output/      JSON and command-facing output assembly
internal/renderer/    Human rendering for TTY, plain text, and markdown
internal/repo/        Local git context and signal correlation
internal/workflow/    Deterministic follow-up workflows
playbooks/bundled/    Starter playbook catalog
docs/                 Architecture, authoring, distribution, and ADRs
```

## Documentation

- [SYSTEM.md](SYSTEM.md): authoritative product and architecture constraints
- [docs/architecture.md](docs/architecture.md): runtime boundaries and catalog resolution
- [docs/playbooks.md](docs/playbooks.md): playbook authoring guidance
- [docs/distribution.md](docs/distribution.md): starter and premium delivery model
- [docs/adr/README.md](docs/adr/README.md): architecture decision records
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md): current implementation status
- [ROADMAP.md](ROADMAP.md): next-step priorities and non-goals

## Development And Validation

Local development:

```bash
make build
make test
make run LOG=build.log
make bench
```

Catalog and release validation:

```bash
make review
make release-check VERSION=v0.1.0
make release-check VERSION=v0.1.0 WITH_DOCKER=1 IMAGE=faultline-smoke
```

When a premium pack is available locally, include cross-pack validation:

```bash
make premium-check PREMIUM_PACK_DIR=../faultline-premium-pack
make premium-review PREMIUM_PACK_DIR=../faultline-premium-pack
make release-check VERSION=v0.1.0 PREMIUM_PACK_DIR=../faultline-premium-pack
```

`make premium-link` creates the ignored local symlink used by the Makefile convenience flow, and `make premium-path` shows which premium directory those targets resolve.
